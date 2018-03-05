package ts

import (
	"context"
	"io"
	"sync"

	"github.com/pkg/errors"
	"github.com/puellanivis/breton/lib/mpeg/ts/packet"
	"github.com/puellanivis/breton/lib/mpeg/ts/psi"
)

type Program struct {
	mu sync.Mutex

	ts *TransportStream

	pid uint16
	pmt *psi.PMT
	wr  io.WriteCloser
}

func (p *Program) PID() uint16 {
	return p.pid
}

func (p *Program) StreamID() uint16 {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pmt == nil {
		return 0
	}

	if p.pmt.Syntax == nil {
		return 0
	}

	return p.pmt.Syntax.TableIDExtension
}

func (p *Program) NewWriter(ctx context.Context, typ ProgramType) (io.WriteCloser, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pmt == nil {
		return nil, errors.New("program is not initialized")
	}

	spid := p.ts.newStreamPID()

	if p.pmt.PCRPID == 0x1FFF {
		p.pmt.PCRPID = spid
	}

	sdata := &psi.StreamData{
		Type: byte(typ),
		PID:  spid,
	}

	p.pmt.Streams = append(p.pmt.Streams, sdata)

	w, err := p.ts.m.WriterByPID(ctx, spid, true)
	if err != nil {
		return nil, err
	}

	if s, ok := w.(*stream); ok {
		s.data = sdata
	}

	return w, nil
}

func (p *Program) StreamPIDs() []uint16 {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pmt == nil {
		return nil
	}

	var streamPIDs []uint16

	for _, s := range p.pmt.Streams {
		streamPIDs = append(streamPIDs, s.PID)
	}

	return streamPIDs
}

func (p *Program) packet(continuity byte) (*packet.Packet, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	b, err := p.pmt.Marshal()
	if err != nil {
		return nil, err
	}

	return &packet.Packet{
		PID:        p.pid,
		PUSI:       true,
		Continuity: continuity & 0x0F,
		Payload:    b,
	}, nil
}
