package ts

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/puellanivis/breton/lib/mpeg/ts/pes"
)

type Program struct {
	pid uint16

	*pes.Reader
	*pes.Writer

	closer func() error
}

func (p *Program) String() string {
	var s string

	switch {
	case p.Reader != nil:
		s = p.Reader.String()

	case p.Writer != nil:
		s = p.Writer.String()
	}

	return fmt.Sprintf("{PID:x%04x %s}", p.pid, s)
}

func (p *Program) Read(b []byte) (n int, err error) {
	if p.Reader == nil {
		return 0, errors.New("program 0x%04x is not open for reading")
	}

	return p.Reader.Read(b)
}

func (p *Program) Write(b []byte) (n int, err error) {
	if p.Writer == nil {
		return 0, errors.New("program 0x%04x is not open for writing")
	}

	return p.Writer.Write(b)
}

func (p *Program) Close() error {
	if p.closer != nil {
		return p.closer()
	}

	return nil
}
