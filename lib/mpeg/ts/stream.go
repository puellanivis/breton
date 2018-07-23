package ts

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/puellanivis/breton/lib/mpeg/ts/psi"
)

// ProgramType defines an enum describing some common MPEG-TS program types.
type ProgramType byte

// ProgramType enum values.
const (
	ProgramTypeVideo ProgramType = 0x01
	ProgramTypeAudio ProgramType = 0x03
	ProgramTypeAAC   ProgramType = 0x0F
	ProgramTypeH264  ProgramType = 0x1B

	ProgramTypeUnknown ProgramType = 0x09 // TODO: this is a guess?
)

type stream struct {
	mu sync.Mutex

	ready         chan struct{}
	discontinuity chan struct{}

	err error

	pid  uint16
	data *psi.StreamData

	rd io.Reader
	wr io.Writer

	closer func() error
}

func (s *stream) makeReady() {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.ready:
	default:
		close(s.ready)
	}
}

func (s *stream) Discontinuity() {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.discontinuity:
	default:
		// already marked, avoid making more channels.
		return
	}

	s.discontinuity = make(chan struct{})
}

func (s *stream) getDiscontinuity() bool {
	select {
	case <-s.discontinuity:
		return false
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.discontinuity == nil {
		s.discontinuity = make(chan struct{})
	}

	select {
	case <-s.discontinuity:
		return false
	default:
	}

	close(s.discontinuity)
	return true
}

func (s *stream) String() string {
	<-s.ready

	out := []string{
		"stream",
		fmt.Sprintf("PID:x%04X", s.pid),
	}

	if s.rd != nil {
		switch s.rd.(type) {
		case fmt.Stringer:
			out = append(out, fmt.Sprintf("R:%v", s.rd))
		default:
			out = append(out, "R")
		}
	}

	if s.wr != nil {
		switch s.wr.(type) {
		case fmt.Stringer:
			out = append(out, fmt.Sprintf("W:%v", s.wr))
		default:
			out = append(out, "W")
		}
	}

	return fmt.Sprintf("{%s}", strings.Join(out, " "))
}

func (s *stream) Read(b []byte) (n int, err error) {
	<-s.ready

	if s.rd == nil {
		if s.err != nil {
			return 0, s.err
		}

		return 0, errors.Errorf("pid 0x%04X is not open for reading", s.pid)
	}

	return s.rd.Read(b)
}

func (s *stream) Write(b []byte) (n int, err error) {
	<-s.ready

	if s.wr == nil {
		if s.err != nil {
			return 0, s.err
		}

		return 0, errors.Errorf("pid 0x%04X is not open for writing", s.pid)
	}

	return s.wr.Write(b)
}

func (s *stream) Close() error {
	s.makeReady()

	if s.closer != nil {
		return s.closer()
	}

	return nil
}
