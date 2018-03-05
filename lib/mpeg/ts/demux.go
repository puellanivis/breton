package ts

import (
	"context"
	"io"
	"sync"

	"github.com/pkg/errors"
	"github.com/puellanivis/breton/lib/glog"
	"github.com/puellanivis/breton/lib/io/bufpipe"
	"github.com/puellanivis/breton/lib/mpeg/ts/packet"
	"github.com/puellanivis/breton/lib/mpeg/ts/pes"
	"github.com/puellanivis/breton/lib/mpeg/ts/psi"
)

var _ = glog.Info

type Demux struct {
	TransportStream

	src io.Reader

	closed chan struct{}

	mu       sync.Mutex
	programs map[uint16]*bufpipe.Pipe

	complete  chan struct{}
	pending   map[uint16]*bufpipe.Pipe
	pendingWG sync.WaitGroup
}

func NewDemux(rd io.Reader, opts ...Option) *Demux {
	d := &Demux{
		src: rd,

		closed: make(chan struct{}),

		programs: make(map[uint16]*bufpipe.Pipe),

		complete: make(chan struct{}),
		pending:  make(map[uint16]*bufpipe.Pipe),
	}

	for _, opt := range opts {
		_ = opt(&d.TransportStream)
	}

	return d
}

const (
	pidPAT  uint16 = 0
	pidNULL uint16 = 0x1FFF
)

func (d *Demux) getPipe(ctx context.Context, pid uint16) (*bufpipe.Pipe, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.programs[pid]; exists {
		return nil, errors.Errorf("pid 0x%04X is already assigned", pid)
	}

	pipe := d.pending[pid]
	if pipe == nil {
		// We assign a context closer below, so don’t assign it here.
		pipe = bufpipe.New(nil, bufpipe.WithNoAutoFlush())
	}
	delete(d.pending, pid)

	// Here we set the context closer for both paths.
	// This way, if a pending pipe were made in the Serve goroutine,
	// we properly tie it to _this_ context, and not the Serve context.
	pipe.CloseOnContext(ctx)

	d.programs[pid] = pipe
	return pipe, nil
}

func (d *Demux) Reader(ctx context.Context, streamID uint16) (io.ReadCloser, error) {
	if streamID == 0 {
		return nil, errors.Errorf("stream_id 0x%04X is invalid", streamID)
	}

	select {
	case <-d.closed:
		return nil, errors.New("Demux is closed")
	default:
	}

	s := &stream{
		ready: make(chan struct{}),
	}

	d.pendingWG.Add(1)
	go func() {
		defer d.pendingWG.Done()
		defer s.makeReady()

		pat := d.GetPAT()

		pmtPID, ok := pat[streamID]
		if !ok {
			s.err = errors.Errorf("no PMT found for stream_id 0x%04X", streamID)
			return
		}

		pmtRD, err := d.ReaderByPID(ctx, pmtPID, false)
		if err != nil {
			s.err = err
			return
		}
		defer pmtRD.Close()

		b := make([]byte, 1024)
		n, err := pmtRD.Read(b)
		if err != nil {
			s.err = err
			return
		}
		b = b[:n]

		if n < 1 {
			s.err = errors.Errorf("zero-length read for pmt on pid 0x%04X", pmtPID)
			return
		}

		tbl, err := psi.Unmarshal(b)
		if err != nil {
			s.err = err
			return
		}

		pmt, ok := tbl.(*psi.PMT)
		if !ok {
			s.err = errors.Errorf("unexpected table on pid 0x%04X: %v", pmtPID, tbl.TableID())
		}

		var pid uint16
		for _, s := range pmt.Streams {
			pid = s.PID
			break
		}

		pipe, err := d.getPipe(ctx, pid)
		if err != nil {
			s.err = err
			return
		}

		s.pid = pid
		s.rd = pes.NewReader(pipe)
		s.closer = func() error {
			d.mu.Lock()
			defer d.mu.Unlock()

			return d.closePID(pid)
		}
	}()

	return s, nil
}

func (d *Demux) ReaderByPID(ctx context.Context, pid uint16, isPES bool) (io.ReadCloser, error) {
	if pid == pidNULL {
		return nil, errors.Errorf("pid 0x%04X is invalid", pid)
	}

	select {
	case <-d.closed:
		return nil, errors.New("Demux is closed")
	default:
	}

	pipe, err := d.getPipe(ctx, pid)
	if err != nil {
		return nil, err
	}

	var rd io.Reader = pipe
	if isPES {
		rd = pes.NewReader(rd)
	}

	ready := make(chan struct{})
	close(ready)

	return &stream{
		ready: ready,

		pid: pid,

		rd: rd,
		closer: func() error {
			d.mu.Lock()
			defer d.mu.Unlock()

			return d.closePID(pid)
		},
	}, nil
}

func (d *Demux) closePending(pid uint16) {
	pipe := d.pending[pid]
	if pipe == nil {
		return
	}

	delete(d.pending, pid)
	pipe.Close()
}

func (d *Demux) closePID(pid uint16) error {
	pipe := d.programs[pid]
	if pipe == nil {
		return nil
	}

	delete(d.programs, pid)
	return pipe.Close()
}

func (d *Demux) Close() <-chan error {
	errch := make(chan error)

	go func() {
		defer close(errch)

		d.mu.Lock()
		defer d.mu.Unlock()

		var pids []uint16
		for pid := range d.programs {
			pids = append(pids, pid)
		}

		for _, pid := range pids {
			if err := d.closePID(pid); err != nil {
				errch <- err
			}
		}

		select {
		case <-d.closed:
		default:
			close(d.closed)
		}
	}()

	return errch
}

func (d *Demux) get(pkt *packet.Packet) (wr *bufpipe.Pipe, debug func(*packet.Packet)) {
	d.mu.Lock()
	defer d.mu.Unlock()

	debug = d.getDebug()

	wr = d.programs[pkt.PID]
	if wr != nil {
		return wr, debug
	}

	select {
	case <-d.complete:
		return nil, debug
	default:
	}

	wr = d.pending[pkt.PID]
	if wr != nil {
		return wr, debug
	}

	// Make a new bufpipe.Pipe with no context closer.
	// A context closer will be attached,
	// only if this is transformed from pending.
	wr = bufpipe.New(nil, bufpipe.WithNoAutoFlush())

	d.pending[pkt.PID] = wr

	return wr, debug
}

func (d *Demux) readOne(b []byte) (error, bool) {
	if _, err := d.src.Read(b); err != nil {
		return err, true
	}

	pkt := new(packet.Packet)
	if err := pkt.Unmarshal(b); err != nil {
		return err, false
	}

	wr, debug := d.get(pkt)

	if debug != nil {
		debug(pkt)
	}

	if wr == nil {
		return nil, false
	}

	if pkt.PUSI {
		if err := wr.Sync(); err != nil {
			d.mu.Lock()
			defer d.mu.Unlock()

			return d.closePID(pkt.PID), false
		}
	}

	if _, err := wr.Write(pkt.Bytes()); err != nil {
		d.mu.Lock()
		defer d.mu.Unlock()

		return d.closePID(pkt.PID), false
	}

	return nil, false
}

func retError(err error) <-chan error {
	errch := make(chan error, 1)
	errch <- err
	close(errch)
	return errch
}

func (d *Demux) Serve(ctx context.Context) <-chan error {
	rdPAT, err := d.ReaderByPID(ctx, pidPAT, false)
	if err != nil {
		return retError(err)
	}

	errch := make(chan error)
	done := make(chan struct{})

	go func() {
		d.pendingWG.Wait()

		close(d.complete)

		d.mu.Lock()
		defer d.mu.Unlock()

		var pids []uint16
		for pid := range d.programs {
			pids = append(pids, pid)
		}

		for _, pid := range pids {
			d.closePending(pid)
		}
	}()

	go func() {
		//defer d.setPAT(nil)

		b := make([]byte, 1024)

		var ver byte = 0xFF

		for {
			n, err := rdPAT.Read(b)
			if err != nil {
				if err != io.EOF {
					select {
					case <-done:
					default:
						errch <- err
					}
				}

				select {
				case <-d.patReady:
					return
				default:
				}

				continue
			}

			if n < 1 {
				// empty-reads are real possibilities.
				continue
			}

			tbl, err := psi.Unmarshal(b)
			if err != nil {
				select {
				case <-done:
				default:
					errch <- err
				}

				select {
				case <-d.patReady:
					return
				default:
				}

				continue
			}

			pat, ok := tbl.(*psi.PAT)
			if !ok {
				errch <- errors.Errorf("unexpected table on pid 0x0000: %v", tbl.TableID())
				continue
			}

			if pat.Syntax != nil {
				if ver == pat.Syntax.Version {
					continue
				}
				ver = pat.Syntax.Version
			}

			newPAT := make(map[uint16]uint16)
			for _, m := range pat.Map {
				newPAT[m.ProgramNumber] = m.PID
			}

			d.setPAT(newPAT)
		}
	}()

	go func() {
		defer func() {
			for err := range d.Close() {
				errch <- err
			}

			close(done)
			close(errch)
		}()

		b := make([]byte, packet.Length)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			err, isFatal := d.readOne(b)
			if err != nil {
				if err == io.EOF {
					return
				}

				errch <- err
			}

			if isFatal {
				return
			}
		}
	}()

	return errch
}
