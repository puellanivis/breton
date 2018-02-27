package ts

import (
	"bytes"
	"context"
	"io"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/puellanivis/breton/lib/glog"
	"github.com/puellanivis/breton/lib/io/bufpipe"
	"github.com/puellanivis/breton/lib/mpeg/ts/pcr"
	"github.com/puellanivis/breton/lib/mpeg/ts/pes"
	"github.com/puellanivis/breton/lib/mpeg/ts/psi"
)

var _ = glog.Info

type sink struct {
	sync.Mutex
	io.Writer

	ticker  chan struct{}
	counter int
}

func (s *sink) Write(b []byte) (n int, err error) {
	s.Lock()
	defer s.Unlock()

	n, err = s.Writer.Write(b)

	s.counter--
	if s.counter <= 0 {
		select {
		case s.ticker <- struct{}{}:
			s.counter += 15 // TODO: make configurable
		default:
		}
	}

	return n, err
}

type pmtDetails struct {
	pid uint16
	pmt *psi.PMT
	wr  io.WriteCloser
}

func (pmt *pmtDetails) marshalPacket(continuity byte) ([]byte, error) {
	b, err := pmt.pmt.Marshal()
	if err != nil {
		return nil, err
	}

	pkt := &Packet{
		PID:        pmt.pid,
		PUSI:       true,
		Continuity: continuity & 0xF,
		Payload:    b,
	}

	return pkt.Marshal()
}

type Mux struct {
	sink sink

	pcrSrc *pcr.Source

	closed chan struct{}
	ready  chan struct{}

	mu          sync.Mutex
	outstanding sync.WaitGroup

	nextStreamPID uint16
	pat           map[uint16]uint16
	pmts          map[uint16]*pmtDetails
}

type MuxOption func(*Mux) MuxOption

/*func WithDebug(fn func(*Packet)) MuxOption {
	return func(m *Mux) MuxOption {
		m.mu.Lock()
		defer m.mu.Unlock()

		save := m.debug
		m.debug = fn

		return WithDebug(save)
	}
}*/

func NewMux(wr io.Writer, opts ...MuxOption) *Mux {
	m := &Mux{
		sink: sink{
			Writer: wr,
			ticker: make(chan struct{}),
		},

		pcrSrc: pcr.NewSource(),

		closed: make(chan struct{}),
		ready:  make(chan struct{}),

		nextStreamPID: 0x100,
	}

	for _, opt := range opts {
		_ = opt(m)
	}

	return m
}

func (m *Mux) assignPID() uint16 {
	m.mu.Lock()
	defer m.mu.Unlock()

	ret := m.nextStreamPID
	m.nextStreamPID++

	return ret
}

func (m *Mux) Writer(ctx context.Context, streamID uint16) (io.WriteCloser, error) {
	if streamID == 0 {
		return nil, errors.Errorf("stream_id 0x%04X is invalid", streamID)
	}

	select {
	case <-m.closed:
		return nil, errors.New("Mux is closed")
	default:
	}

	pid := m.assignPID()
	pmtPID := pid << 4

	wr, err := m.WriterByPID(ctx, pmtPID, false)
	if err != nil {
		return nil, err
	}

	pmt := &pmtDetails{
		pid: pmtPID,
		pmt: &psi.PMT{
			Syntax: &psi.SectionSyntax{
				TableIDExtension: streamID,
				Current:          true,
			},
			PCRPID: pid,
			Streams: []*psi.StreamData{
				&psi.StreamData{
					Type: 0x03, // TODO: don’t hardcode audio like this.
					PID:  pid,
				},
			},
		},
		wr: wr,
	}

	if m.pat == nil {
		m.pat = make(map[uint16]uint16)
	}

	m.pat[streamID] = pmtPID

	if m.pmts == nil {
		m.pmts = make(map[uint16]*pmtDetails)
	}

	m.pmts[pid] = pmt

	return m.WriterByPID(ctx, pid, true)
}

const (
	maxLengthAllowingStuffing = packetMaxPayload - adaptationFieldMinLength
)

func (m *Mux) packetizer(pid uint16, isPES bool, rd io.ReadCloser) {
	defer func() {
		if err := rd.Close(); err != nil {
			glog.Error("packetizer: rd.Close: %+v", err)
		}
	}()

	var continuity byte
	discontinuity := false // TODO: make this configurable.

	// PSI table length is limited to 1021 bytes. This is significantly less than 0x10000 bytes.
	// PES packet limited to payload length 0xFFFF, but a header of at least 6, so must be > 0x10000 bytes.
	// So, we use 0x20000 bytes just to be sure we get a whole packet sequence.
	//
	// N.B. It is the Write/Reader’s responsibility to ensure that a Read completes only on full packets,
	// and that said packet sequence will not be > 0x20000 bytes.
	buf := make([]byte, 0x20000)

	for {
		n, err := rd.Read(buf)
		if err != nil {
			if err != io.EOF {
				glog.Errorf("packetizer : 0x%04x : %+v", pid, err)
			}

			return
		}

		// trunc the buffer to only what was read.
		data := buf[:n]

		pusi := true

		for len(data) > 0 {
			var af *AdaptationField
			var l int

			switch {
			case !isPES:
				l = len(data)
				// don’t do anything

			case pusi:
				af = &AdaptationField{
					Discontinuity: discontinuity,
					RandomAccess:  true, // TODO: make this configurable.
					PCR:           new(pcr.PCR),
				}

				m.pcrSrc.Read(af.PCR)

				discontinuity = false

			case len(data) < maxLengthAllowingStuffing:
				// If the remaining payload is small enough to add stuffing and finish this sequence.
				af = &AdaptationField{
					Stuffing: maxLengthAllowingStuffing - len(data),
				}

			case len(data) < packetMaxPayload:
				// We don’t have enough room to add stuffing and finish this sequence.
				// So, we add an empty AdaptationField here with 0-bytes of stuffing,
				// which adds 2-bytes to the header, and overflows the last byte
				// of payload into the next packet,  where we will surely have enough room
				// to actually add stuffing.
				// TODO: check if we can just say adaptation_field_length is 0, which would add only one-byte instead of two?
				af = &AdaptationField{}
			}

			if isPES {
				l = packetMaxPayload - af.len()

				if l > len(data) {
					glog.Errorf("calculated bad payload length: %d > %d", l, len(data))
				}
			}

			pkt := &Packet{
				PID:             pid,
				PUSI:            pusi,
				Continuity:      continuity,
				AdaptationField: af,
				Payload:         data[:l],
			}

			pusi = false
			continuity = (continuity + 1) & 0x0F

			b, err := pkt.Marshal()
			if err != nil {
				glog.Errorf("%+v", err)
				return
			}

			if len(b) != packetLength {
				panic("packet marshaled to size other than 188")
			}

			if _, err := m.sink.Write(b); err != nil {
				glog.Errorf("m.sink.Write: 0x%04x: %+v", pid, err)
				return
			}

			data = data[l:]
		}
	}

}

func (m *Mux) WriterByPID(ctx context.Context, pid uint16, isPES bool) (io.WriteCloser, error) {
	glog.Infof("pid:x%04x, isPES:%v", pid, isPES)

	if pid == pidNULL {
		return nil, errors.Errorf("pid 0x%04X is invalid", pid)
	}

	select {
	case <-m.closed:
		return nil, errors.New("Mux is closed")
	default:
	}

	ready := make(chan struct{})
	close(ready)

	ctx, cancel := context.WithCancel(ctx)
	pipe := bufpipe.New(ctx)
	var rd io.ReadCloser = pipe
	// if !isPES: pipe -> ReadAll -> Packetize -> Marshal -> sink.Write

	if isPES {
		// We only want to wg.Wait on PES streams.
		m.outstanding.Add(1)

		var wr io.WriteCloser

		rd, wr = io.Pipe() // synchronous pipe, don’t write to it without a Reader available.

		pesWR := pes.NewWriter(0xC0, wr) // TODO: don’t hardcode a value for audio.
		//pesWR.Stream.Header.PTS = new(uint64) // we would need to extract this from the input stream…

		// 176      : first payload size minus len(AF{PCR:xxx})
		// - 9      : PES header size
		// 14 * 184 : 14 packets of full payload
		// 182      : 1 packet with enough room for len(AF{Stuffing:xxx})
		bufpipe.WithMaxOutstanding(176 - 9 + 14*184 + 182)(pipe) // TODO: don’t hard code thise.
		bufpipe.WithNoAutoFlush()(pipe)

		go func() {
			defer wr.Close()

			for {
				data, err := pipe.ReadAll()
				if err != nil {
					if err != io.EOF {
						glog.Errorf("pipe.ReadAll: %+v", err)
					}
					return
				}

				if _, err := pesWR.Write(data); err != nil {
					glog.Errorf("mpeg/ts/pes.Writer: %+v", err)
					return
				}
			}
		}()
	}

	go func() {
		if isPES {
			defer m.outstanding.Done()

			// Here, we wait until we’ve written the initial PAT and PMTs.
			<-m.ready
		}

		m.packetizer(pid, isPES, rd)
	}()

	return &program{
		ready: ready,

		pid: pid,

		wr: pipe,
		closer: func() error {
			cancel()

			return pipe.Close()
		},
	}, nil
}

func (m *Mux) Close() <-chan error {
	errch := make(chan error)

	go func() {
		defer close(errch)

		close(m.closed)

		m.outstanding.Wait()
	}()

	return errch
}

func (m *Mux) markReady() {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-m.ready:
	default:
		close(m.ready)
	}
}

func (m *Mux) getPAT() map[uint16]uint16 {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.pat
}

func (m *Mux) getPMTs() map[uint16]*pmtDetails {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.pmts
}

func (m *Mux) writePAT(w io.Writer) error {
	pat := m.getPAT()

	var keys []uint16
	for key := range pat {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	tbl := &psi.PAT{
		Syntax: &psi.SectionSyntax{
			TableIDExtension: 0x1,
			Current:          true,
		},
		Map: make([]psi.ProgramMap, len(keys)),
	}

	for i, key := range keys {
		tbl.Map[i].Set(key, pat[key])
	}

	b, err := tbl.Marshal()
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	return err
}

func (m *Mux) preamble(continuity byte) error {
	preamble := new(bytes.Buffer)

	if err := m.writePAT(preamble); err != nil {
		return err
	}

	pkt := &Packet{
		PID:        pidPAT,
		PUSI:       true,
		Continuity: continuity & 0x0F,
		Payload:    preamble.Bytes(),
	}

	b, err := pkt.Marshal()
	if err != nil {
		return err
	}

	m.sink.Lock()
	defer m.sink.Unlock()

	m.sink.counter += 15

	if _, err := m.sink.Writer.Write(b); err != nil {
		return err
	}

	for _, pmt := range m.getPMTs() {
		b, err := pmt.marshalPacket(continuity)
		if err != nil {
			return err
		}

		_, err = m.sink.Writer.Write(b)
		if err != nil {
			return err
		}
	}

	m.markReady()

	return nil
}

func (m *Mux) Serve(ctx context.Context) <-chan error {
	wrPAT, err := m.WriterByPID(ctx, pidPAT, false)
	if err != nil {
		return retError(err)
	}

	var continuity byte
	if err := m.preamble(continuity); err != nil {
		return retError(err)
	}
	continuity++

	errch := make(chan error)

	go func() {
		defer func() {
			if err := wrPAT.Close(); err != nil {
				errch <- err
			}

			close(errch)
		}()

		// TODO: what do the specifications say?
		timer := time.NewTimer(5 * time.Millisecond)
		defer timer.Stop()

		for {
			timer.Reset(5 * time.Millisecond)

			select {
			case <-ctx.Done():
				return
			case <-m.closed:
				return
			case <-timer.C:
			case <-m.sink.ticker:
			}

			if err := m.preamble(continuity); err != nil {
				errch <- err
				return
			}
			continuity++
		}
	}()

	return errch
}
