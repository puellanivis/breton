// Package ts implements a Mux/Demux for the MPEG Transport Stream protocol.
package ts

import (
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/puellanivis/breton/lib/mpeg/ts/dvb"
	"github.com/puellanivis/breton/lib/mpeg/ts/packet"
	"github.com/puellanivis/breton/lib/mpeg/ts/psi"
)

// PacketSize is the length in bytes of an MPEG-TS packet.
const PacketSize = packet.Length

// Option defines a function that will apply some value or behavior to a TransportStream.
type Option func(*TransportStream) Option

// WithDebug sets a function that will be called for each MPEG-TS Packet processed by the TransportStream.
func WithDebug(fn func(*packet.Packet)) Option {
	return func(ts *TransportStream) Option {
		ts.once.Do(ts.init)
		ts.mu.Lock()
		defer ts.mu.Unlock()

		save := ts.debug
		ts.debug = fn

		return WithDebug(save)
	}
}

func (ts *TransportStream) getDebug() func(*packet.Packet) {
	ts.once.Do(ts.init)
	ts.mu.Lock()
	defer ts.mu.Unlock()

	return ts.debug
}

// WithUpdateRate sets the rate at which the TransportStream will update TODO.
func WithUpdateRate(d time.Duration) Option {
	return func(ts *TransportStream) Option {
		ts.once.Do(ts.init)
		ts.mu.Lock()
		defer ts.mu.Unlock()

		save := ts.updateRate
		ts.updateRate = d

		return WithUpdateRate(save)
	}
}

func (ts *TransportStream) getUpdateRate() time.Duration {
	ts.once.Do(ts.init)
	ts.mu.Lock()
	defer ts.mu.Unlock()

	return ts.updateRate
}

// TransportStream defines an MPEG Transport Stream.
type TransportStream struct {
	sink io.Writer

	ticker  chan struct{}
	counter int

	once sync.Once
	mu   sync.Mutex

	debug      func(*packet.Packet)
	updateRate time.Duration

	patReady chan struct{}
	pat      map[uint16]uint16
	pmts     map[uint16]*Program
	dvbSDT   *dvb.ServiceDescriptorTable

	nextStreamPID  uint16
	nextProgramPID uint16
	lastStreamID   uint16

	m *Mux
}

func (ts *TransportStream) init() {
	ts.patReady = make(chan struct{})
	ts.nextStreamPID = 0x100
	ts.nextProgramPID = 0x1000
	ts.updateRate = 1 * time.Second / 25 // 25 Hz
}

func (ts *TransportStream) writePackets(pkts ...*packet.Packet) (n int, err error) {
	debug := ts.getDebug()

	var q [][]byte

	for _, pkt := range pkts {
		if debug != nil {
			debug(pkt)
		}

		b, err2 := pkt.Marshal()
		if err2 != nil {
			if err == nil {
				err = err2
			}

			continue
		}

		q = append(q, b)
	}

	if len(q) < 1 {
		return 0, err
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()

	for _, b := range q {
		n2, err2 := ts.sink.Write(b)

		n += n2
		if err == nil {
			err = err2
		}
	}

	if len(q) > 1 {
		ts.counter += 15
		return n, err
	}

	ts.counter--
	if ts.counter <= 0 {
		select {
		case ts.ticker <- struct{}{}:
			ts.counter += 15 // TODO: make configurable
		default:
		}
	}

	return n, err
}

func (ts *TransportStream) newStreamPID() uint16 {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	pid := ts.nextStreamPID
	ts.nextStreamPID++
	return pid
}

// NewProgram returns a new Program assigned to the given Stream ID.
func (ts *TransportStream) NewProgram(streamID uint16) (*Program, error) {
	ts.once.Do(ts.init)
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.m == nil {
		return nil, errors.New("cannot make a new program on a non-outbound transport stream")
	}

	if streamID == 0 {
		streamID = ts.lastStreamID + 1

		for ts.pat[streamID] != 0 {
			streamID++
		}
	}
	ts.lastStreamID = streamID

	if ts.pat[streamID] != 0 {
		return nil, errors.Errorf("stream_id 0x%04X is already assigned", streamID)
	}

	pid := ts.nextProgramPID
	ts.nextProgramPID++

	p := &Program{
		pid: pid,
		ts:  ts,
		pmt: &psi.PMT{
			Syntax: &psi.SectionSyntax{
				TableIDExtension: streamID,
				Current:          true,
			},
			PCRPID: 0x1FFF,
		},
	}

	if ts.pat == nil {
		ts.pat = make(map[uint16]uint16)

		select {
		case <-ts.patReady:
		default:
			close(ts.patReady)
		}
	}

	ts.pat[streamID] = pid

	if ts.pmts == nil {
		ts.pmts = make(map[uint16]*Program)
	}

	ts.pmts[pid] = p

	return p, nil
}

func cloneSDT(src *dvb.ServiceDescriptorTable) *dvb.ServiceDescriptorTable {
	if src == nil {
		return nil
	}

	sdt := *src

	if src.Syntax != nil {
		s := *src.Syntax
		sdt.Syntax = &s
	}

	sdt.Services = nil
	for _, s := range src.Services {
		n := *s
		n.Descriptors = nil

		for _, d := range s.Descriptors {
			if d, ok := d.(*dvb.ServiceDescriptor); ok {
				nd := *d
				n.Descriptors = append(n.Descriptors, &nd)
			}
		}

		sdt.Services = append(sdt.Services, &n)
	}

	return &sdt
}

// SetDVBSDT sets the DVB Service Description Table for the TransportStream.
func (ts *TransportStream) SetDVBSDT(sdt *dvb.ServiceDescriptorTable) {
	clone := cloneSDT(sdt)

	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.dvbSDT = clone
}

func (ts *TransportStream) getDVBSDT() *dvb.ServiceDescriptorTable {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	return ts.dvbSDT
}

// GetDVBSDT returns a deep copy of the DVB Service Description Table being used by the TransportStream.
func (ts *TransportStream) GetDVBSDT() *dvb.ServiceDescriptorTable {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	return cloneSDT(ts.dvbSDT)
}

// GetPAT returns a map defining the Program Allocation Table of the TransportStream.
func (ts *TransportStream) GetPAT() map[uint16]uint16 {
	ts.once.Do(ts.init)

	<-ts.patReady

	ts.mu.Lock()
	defer ts.mu.Unlock()

	return ts.pat
}

func (ts *TransportStream) setPAT(pat map[uint16]uint16) {
	ts.once.Do(ts.init)
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if pat != nil {
		ts.pat = pat
	}

	select {
	case <-ts.patReady:
	default:
		close(ts.patReady)
	}
}

// GetPMTs returns a map defining the Program Map Tables indexed by their pid.
//
// TODO: confirm behavior.
func (ts *TransportStream) GetPMTs() map[uint16]*Program {
	ts.once.Do(ts.init)
	ts.mu.Lock()
	defer ts.mu.Unlock()

	return ts.pmts
}
