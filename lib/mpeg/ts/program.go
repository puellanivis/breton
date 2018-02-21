package ts

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

type program struct {
	mu    sync.Mutex
	ready chan struct{}

	err error

	pid uint16

	rd io.Reader
	wr io.Writer

	closer func() error
}

func (p *program) makeReady() {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.ready:
	default:
		close(p.ready)
	}
}

func (p *program) String() string {
	<-p.ready

	out := []string{
		fmt.Sprintf("PID:x%04X", p.pid),
	}

	if p.rd != nil {
		switch p.rd.(type) {
		case fmt.Stringer:
			out = append(out, fmt.Sprintf("R:%v", p.rd))
		default:
			out = append(out, "R")
		}
	}

	if p.wr != nil {
		switch p.wr.(type) {
		case fmt.Stringer:
			out = append(out, fmt.Sprintf("W:%v", p.wr))
		default:
			out = append(out, "W")
		}
	}

	return fmt.Sprintf("{%s}", strings.Join(out, " "))
}

func (p *program) Read(b []byte) (n int, err error) {
	<-p.ready

	if p.rd == nil {
		if p.err != nil {
			return 0, p.err
		}

		return 0, errors.Errorf("program pid 0x%04X is not open for reading", p.pid)
	}

	return p.rd.Read(b)
}

func (p *program) Write(b []byte) (n int, err error) {
	<-p.ready

	if p.wr == nil {
		if p.err != nil {
			return 0, p.err
		}

		return 0, errors.Errorf("program pid 0x%04X is not open for writing", p.pid)
	}

	return p.wr.Write(b)
}

func (p *program) Close() error {
	p.makeReady()

	if p.closer != nil {
		return p.closer()
	}

	return nil
}
