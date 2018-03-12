package ts

import (
	"context"
	"io"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/puellanivis/breton/lib/glog"
	"github.com/puellanivis/breton/lib/io/bufpipe"
	"github.com/puellanivis/breton/lib/mpeg/ts/dvb"
	"github.com/puellanivis/breton/lib/mpeg/ts/packet"
	"github.com/puellanivis/breton/lib/mpeg/ts/pcr"
	"github.com/puellanivis/breton/lib/mpeg/ts/pes"
	"github.com/puellanivis/breton/lib/mpeg/ts/psi"
)

var _ = glog.Info

type Mux struct {
	TransportStream

	pcrSrc *pcr.Source

	closed chan struct{}
	ready  chan struct{}
	chain  chan struct{} // chain is used to ensure that the first packet of each stream is in order.

	mu          sync.Mutex
	outstanding sync.WaitGroup
}

func NewMux(wr io.Writer, opts ...Option) *Mux {
	chain := make(chan struct{})
	close(chain)

	m := &Mux{
		TransportStream: TransportStream{
			sink:   wr,
			ticker: make(chan struct{}),
		},

		pcrSrc: pcr.NewSource(),

		closed: make(chan struct{}),
		ready:  make(chan struct{}),
		chain:  chain,
	}

	m.TransportStream.m = m

	for _, opt := range opts {
		_ = opt(&m.TransportStream)
	}

	return m
}

func (m *Mux) NewProgram(ctx context.Context, streamID uint16) (*Program, error) {
	if streamID == 0 {
		return nil, errors.Errorf("stream_id 0x%04X is invalid", streamID)
	}

	select {
	case <-m.closed:
		return nil, errors.New("Mux is closed")
	default:
	}

	p, err := m.TransportStream.NewProgram(streamID)
	if err != nil {
		return nil, err
	}

	wr, err := m.WriterByPID(ctx, p.PID(), false)
	if err != nil {
		return nil, err
	}
	p.wr = wr

	return p, nil
}

func (m *Mux) Writer(ctx context.Context, streamID uint16, typ ProgramType) (io.WriteCloser, error) {
	p, err := m.NewProgram(ctx, streamID)
	if err != nil {
		return nil, err
	}

	return p.NewWriter(ctx, typ)
}

const (
	maxLengthAllowingStuffing = packet.MaxPayload - packet.AdaptationFieldMinLength
)

func (m *Mux) packetizer(rd io.ReadCloser, s *stream, isPES bool, next chan struct{}) {
	defer func() {
		if err := rd.Close(); err != nil {
			glog.Error("packetizer: 0x%04X: rd.Close: %+v", s.pid, err)
		}
	}()

	var continuity byte

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
				glog.Errorf("packetizer: 0x%04X: %+v", s.pid, err)
			}

			return
		}

		if n == len(buf) {
			glog.Warningf("packetizer: 0x%04X: unexpected full read of packet buffer")
		}

		// trunc the buffer to only what was read.
		data := buf[:n]

		pusi := true

		for len(data) > 0 {
			var af *packet.AdaptationField

			switch {
			case !isPES:
				// Don’t do anything, PSI tables don’t get an AdaptationField.

			case pusi:
				af = &packet.AdaptationField{
					Discontinuity: s.getDiscontinuity(),
					RandomAccess:  true, // TODO: make this configurable.
					PCR:           new(pcr.PCR),
				}

				m.pcrSrc.Read(af.PCR)

			case len(data) < maxLengthAllowingStuffing:
				// If the remaining payload is small enough to add stuffing and finish this sequence.
				af = &packet.AdaptationField{
					Stuffing: maxLengthAllowingStuffing - len(data),
				}

			case len(data) < packet.MaxPayload:
				// We don’t have enough room to add stuffing and finish this sequence.
				// So, we add an empty AdaptationField here with 0-bytes of stuffing,
				// which adds 2-bytes to the header, and overflows the last byte
				// of payload into the next packet,  where we will surely have enough room
				// to actually add stuffing.
				// TODO: check if we can just say adaptation_field_length is 0, which would add only one-byte instead of two?
				af = &packet.AdaptationField{}
			}

			l := packet.MaxPayload - af.Len()

			if l > len(data) {
				switch {
				case isPES:
					if af != nil {
						af.Stuffing -= len(data) - l
						l = packet.MaxPayload - af.Len()
					}

					if l > len(data) {
						glog.Errorf("calculated bad payload space: %d > %d", l, len(data))
					}

				default:
					l = len(data)
				}
			}

			pkt := &packet.Packet{
				PID:             s.pid,
				PUSI:            pusi,
				Continuity:      continuity,
				AdaptationField: af,
				Payload:         data[:l],
			}

			pusi = false
			continuity = (continuity + 1) & 0x0F

			if _, err := m.writePackets(pkt); err != nil {
				glog.Errorf("m.writePackets: 0x%04X: %+v", s.pid, err)
				return
			}

			if next != nil {
				close(next)
				next = nil
			}

			data = data[l:]
		}
	}

}

func (m *Mux) chainLink() (<-chan struct{}, chan struct{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	wait, next := m.chain, make(chan struct{})
	m.chain = next

	return wait, next
}

func (m *Mux) WriterByPID(ctx context.Context, pid uint16, isPES bool) (io.WriteCloser, error) {
	glog.Infof("pid:x%04X, isPES:%v", pid, isPES)

	if pid == pidNULL {
		return nil, errors.Errorf("pid 0x%04X is invalid", pid)
	}

	select {
	case <-m.closed:
		return nil, errors.Errorf("pid 0x%04X: mux is closed", pid)
	default:
	}

	pipe := bufpipe.New(ctx)

	var rd io.ReadCloser = pipe
	// if !isPES: bufpipe.Pipe -> Packetizer
	// if  isPES: bufpipe.Pipe -> ReadAll -> pes.Writer -> io.Pipe -> Packetizer

	var wait <-chan struct{}
	var next chan struct{}

	if isPES {
		wait, next = m.chainLink()

		var wr io.WriteCloser

		rd, wr = io.Pipe() // synchronous pipe, don’t write to it without a Reader available.

		pesWR := pes.NewWriter(0xC0, wr) // TODO: don’t hardcode a value for audio.
		//pesWR.Stream.Header.PTS = new(uint64) // we would need to extract this from the input stream…

		pesHdrLen, err := pesWR.HeaderLength()
		if err != nil {
			return nil, err
		}

		// 176                       : first payload size (MaxPayload - len(AF{PCR:xxx}))
		// pes.HeaderLength          : PES header size
		// 14 * packet.MaxPayload    : 14 packets of full payload
		// maxLengthAllowingStuffing : 1 packet with enough room for len(AF{Stuffing:xxx})
		// = |PUSI:payload[176]|, 14 × |payload[184]|, |AF{Stuffing}:payload[<182]|

		// TODO: don’t hard code thise.
		bufpipe.WithMaxOutstanding(176 - pesHdrLen + 14*packet.MaxPayload + maxLengthAllowingStuffing)(pipe)
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
					glog.Errorf("mpeg/ts/pes.Writer.Write: %+v", err)
					return
				}
			}
		}()
	}

	if isPES {
		// We only want to wg.Wait on PES streams.
		m.outstanding.Add(1)
	}

	ready := make(chan struct{})
	close(ready)

	s := &stream{
		ready: ready,

		pid: pid,

		wr: pipe,
		closer: func() error {
			return pipe.Close()
		},
	}

	go func() {
		if isPES {
			defer m.outstanding.Done()

			// Here, we wait until we’ve written the initial PAT and PMTs.
			<-m.ready

			// Here, we wait for our turn in the chain,
			// to ensure a deterministic order of first packets.
			<-wait
		}

		m.packetizer(rd, s, isPES, next)
	}()

	return s, nil
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

func (m *Mux) marshalPAT() ([]byte, error) {
	pat := m.GetPAT()

	var keys []uint16
	for key := range pat {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	pmap := make([]psi.ProgramMap, len(keys))
	for i, key := range keys {
		pmap[i].Set(key, pat[key])
	}

	tbl := &psi.PAT{
		Syntax: &psi.SectionSyntax{
			TableIDExtension: 0x1,
			Current:          true,
		},
		Map: pmap,
	}

	return tbl.Marshal()
}

// this needs to be moved into sink, as right now it violates sink’s internal details.
func (m *Mux) preamble(continuity byte) error {
	var pkts []*packet.Packet

	continuity = continuity & 0x0F

	if sdt := m.TransportStream.getDVBSDT(); sdt != nil {
		if payload, err := sdt.Marshal(); err == nil {
			// In this specific case, if we get an error, just ignore the packet entirely.
			pkts = append(pkts, &packet.Packet{
				PID:        dvb.ServiceDescriptionPID,
				PUSI:       true,
				Continuity: continuity,
				Payload:    payload,
			})
		} else {
			glog.Warningf("dvb.ServiceDescriptorTable.Marshal: %+v", err)
		}
	}

	payload, err := m.marshalPAT()
	if err != nil {
		return err
	}

	pkts = append(pkts, &packet.Packet{
		PID:        pidPAT,
		PUSI:       true,
		Continuity: continuity,
		Payload:    payload,
	})

	for _, p := range m.GetPMTs() {
		pkt, err := p.packet(continuity)
		if err != nil {
			return err
		}

		pkts = append(pkts, pkt)
	}

	if _, err := m.writePackets(pkts...); err != nil {
		return err
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
		timer := time.NewTimer(m.getUpdateRate())
		defer timer.Stop()

		for {
			timer.Reset(m.getUpdateRate())

			select {
			case <-ctx.Done():
				return
			case <-m.closed:
				return
			case <-timer.C:
			case <-m.ticker:
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
