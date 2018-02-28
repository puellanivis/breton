// Packet ts implements a Mux/Demux for the MPEG Transport Stream protocol.
package ts

import (
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/puellanivis/breton/lib/mpeg/ts/packet"
	"github.com/puellanivis/breton/lib/mpeg/ts/psi"
)

const PacketSize = packet.Length

type Option func(*TransportStream) Option

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
	pmts     map[uint16]*ProgramDetails

	nextStreamPID uint16
	lastStreamID  uint16
}

func (ts *TransportStream) init() {
	ts.patReady = make(chan struct{})
	ts.nextStreamPID = 0x100
	ts.updateRate = 5 * time.Millisecond
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

func (ts *TransportStream) NewProgram(streamID uint16, typ ProgramType) (*ProgramDetails, error) {
	ts.once.Do(ts.init)
	ts.mu.Lock()
	defer ts.mu.Unlock()

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

	pid := ts.nextStreamPID
	ts.nextStreamPID++

	pmtPID := pid << 4

	pd := &ProgramDetails{
		pid: pmtPID,
		pmt: &psi.PMT{
			Syntax: &psi.SectionSyntax{
				TableIDExtension: streamID,
				Current:          true,
			},
			PCRPID: pid,
			Streams: []*psi.StreamData{
				&psi.StreamData{
					Type: byte(typ),
					PID:  pid,
				},
			},
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

	ts.pat[streamID] = pmtPID

	if ts.pmts == nil {
		ts.pmts = make(map[uint16]*ProgramDetails)
	}

	ts.pmts[pid] = pd

	return pd, nil
}

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

func (ts *TransportStream) GetPMTs() map[uint16]*ProgramDetails {
	ts.once.Do(ts.init)
	ts.mu.Lock()
	defer ts.mu.Unlock()

	return ts.pmts
}
