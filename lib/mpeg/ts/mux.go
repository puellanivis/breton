package ts

import (
	"context"
	"io"
	"sync"

	"github.com/pkg/errors"
	"github.com/puellanivis/breton/lib/glog"
	"github.com/puellanivis/breton/lib/io/bufpipe"
	"github.com/puellanivis/breton/lib/mpeg/ts/pcr"
	"github.com/puellanivis/breton/lib/mpeg/ts/pes"
)

var _ = glog.Info

type sink struct {
	sync.Mutex
	io.Writer
}

func (s *sink) Write(b []byte) (n int, err error) {
	s.Lock()
	defer s.Unlock()

	return s.Writer.Write(b)
}

type Mux struct {
	sink sink

	pcrSrc *pcr.Source

	closed chan struct{}

	mu sync.Mutex
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
		},

		pcrSrc: pcr.NewSource(),

		closed:   make(chan struct{}),
	}

	for _, opt := range opts {
		_ = opt(m)
	}

	return m
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

	return nil, errors.New("unimplemented")
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

	if isPES {
		var wr io.WriteCloser

		rd, wr = io.Pipe() // synchronous pipe, don’t write to it without a Reader available.

		pesWR := pes.NewWriter(0xC0, wr) // TODO: don’t hardcode this
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
		defer rd.Close()

		var continuity byte
		discontinuity := false

		buf := make([]byte, 0xffff)

		for {
			n, err := rd.Read(buf)
			if err != nil {
				if err != io.EOF {
					glog.Errorf("%d: %+v", n, err)
				}

				return
			}

			// trunc the buffer to only what was read.
			data := buf[:n]

			PUSI := true

			for len(data) > 0 {
				var af *AdaptationField

				switch {
				case PUSI:
					af = &AdaptationField{
						Discontinuity: discontinuity,	// TODO: make initial state optional
						RandomAccess: true,		// TODO: make optional
						PCR: new(pcr.PCR),
					}

					m.pcrSrc.Read(af.PCR)

					discontinuity = false

				case len(data) < 182: // TODO: make not a magic number
					af = &AdaptationField{
						Stuffing: 182 - len(data), // TODO: make not a magic number
					}

				case len(data) < 184:
					// We don’t have enough room to add stuffing and finish this whole sequence.
					// So, we add an AdaptationField here with 0-bytes of stuff,
					// which will add 2-bytes to the header,
					// and overflow the last two bytes of payload into the next packet,
					// where we will have enough room to actually add stuffing.
					af = &AdaptationField{}
				}

				l := 184 - af.len() // TODO: make not a magic number.

				if l > len(data) {
					glog.Errorf("calculated bad payload length: %d > %d", l, len(data))
				}

				pkt := &Packet{
					PID: pid,
					PUSI: PUSI,
					Continuity: continuity,
					AdaptationField: af,
					payload: data[:l],
				}

				PUSI = false
				continuity = (continuity + 1) & 0x0F


				b, err := pkt.Marshal()
				if err != nil {
					glog.Errorf("%+v", err)
					return
				}

				if len(b) != 188 {
					panic("packet marshaled to size other than 188")
				}

				if _, err := m.sink.Write(b); err != nil {
					glog.Errorf("%+v", err)
					return
				}

				data = data[l:]
			}
		}
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

		m.mu.Lock()
		defer m.mu.Unlock()

	}()

	return errch
}

func (m *Mux) Serve(ctx context.Context) <-chan error {
	/*wrPAT, err := m.WriterByPID(ctx, pidPAT, false)
	if err != nil {
		return retError(err)
	} //*/

	errch := make(chan error)

	go func() {
		defer close(errch)
		//defer wrPAT.Close()

		<-ctx.Done()
	}()

	return errch
}
