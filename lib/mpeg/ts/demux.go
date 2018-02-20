package ts

import (
	"context"
	"io"
	"sync"

	"github.com/pkg/errors"
	"github.com/puellanivis/breton/lib/io/bufpipe"
	"github.com/puellanivis/breton/lib/mpeg/ts/pes"
	"github.com/puellanivis/breton/lib/mpeg/ts/psi"
)

type Demux struct {
	src io.Reader

	closed chan struct{}
	debug  func(*Packet)

	mu       sync.Mutex
	programs map[uint16]*bufpipe.Pipe
	pat      map[uint16]uint16
}

type DemuxOption func(*Demux) DemuxOption

func WithDebug(fn func(*Packet)) DemuxOption {
	return func(d *Demux) DemuxOption {
		d.mu.Lock()
		defer d.mu.Unlock()

		save := d.debug
		d.debug = fn

		return WithDebug(save)
	}
}

func NewDemux(rd io.Reader, opts ...DemuxOption) *Demux {
	d := &Demux{
		src:    rd,
		closed: make(chan struct{}),

		programs: make(map[uint16]*bufpipe.Pipe),
	}

	for _, opt := range opts {
		_ = opt(d)
	}

	return d
}

const (
	pidPAT  uint16 = 0
	pidNULL uint16 = 0x1FFF
)

func (d *Demux) ReaderByPID(ctx context.Context, pid uint16) (*Program, error) {
	if pid == pidNULL {
		return nil, errors.Errorf("pid 0x%04x is invalid", pid)
	}

	select {
	case <-d.closed:
		return nil, errors.New("Demux is closed")
	default:
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.programs[pid]; exists {
		return nil, errors.Errorf("pid 0x%04x is already assigned", pid)
	}

	pipe := bufpipe.New(ctx)
	d.programs[pid] = pipe

	return &Program{
		pid:    pid,
		Reader: pes.NewReader(pipe),
		closer: func() error {
			d.mu.Lock()
			defer d.mu.Unlock()

			return d.closePID(pid)
		},
	}, nil
}

func (d *Demux) closePID(pid uint16) error {
	pipe := d.programs[pid]
	if pipe == nil {
		return nil
	}

	d.programs[pid] = nil
	return pipe.Close()
}

func (d *Demux) Close() <-chan error {
	errch := make(chan error)

	go func() {
		defer close(errch)

		d.mu.Lock()
		defer d.mu.Unlock()

		for pid, _ := range d.programs {
			if err := d.closePID(pid); err != nil {
				errch <- err
			}
		}
	}()

	return errch
}

func (d *Demux) get(pkt *Packet) (wr *bufpipe.Pipe, debug func(*Packet)) {
	var newPAT map[uint16]uint16

	if pat, ok := pkt.PSI.(*psi.PAT); ok {
		newPAT = make(map[uint16]uint16)

		for _, m := range pat.Map {
			newPAT[m.ProgramNumber] = m.PID
		}
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if newPAT != nil {
		d.pat = newPAT
	}

	return d.programs[pkt.PID], d.debug
}

func (d *Demux) readOne(b []byte) (error, bool) {
	if _, err := d.src.Read(b); err != nil {
		return err, true
	}

	pkt := new(Packet)
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

	if _, err := wr.Write(pkt.Bytes()); err != nil {
		d.mu.Lock()
		defer d.mu.Unlock()

		return d.closePID(pkt.PID), false
	}

	return nil, false
}

func (d *Demux) Serve(ctx context.Context) <-chan error {
	errch := make(chan error)

	go func() {
		defer func() {
			for err := range d.Close() {
				errch <- err
			}

			close(errch)
		}()

		b := make([]byte, pktLen)

		for {
			select {
			case <-ctx.Done():
				return

			default:
			}

			err, done := d.readOne(b)
			if err != nil {
				if err == io.EOF {
					return
				}

				errch <- err
			}

			if done {
				return
			}
		}
	}()

	return errch
}
