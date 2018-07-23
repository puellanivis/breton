package pcr

import (
	"time"
)

// Source represents a source of PCR values.
type Source struct {
	t time.Time
}

// NewSource establishes a new PCR source that counts from the current time as of the this call.
func NewSource() *Source {
	return &Source{
		t: time.Now(),
	}
}

// Read sets the given PCR from this Source.
func (s *Source) Read(pcr *PCR) {
	// There is a possible discontinuity due to math overflow at (time.Duration(1 << 64) / 27)
	// However, math says that should be at about ~21.65 years on average.
	// If a stream is running for that long, this code could go haywire.
	pcr.Set(time.Since(s.t))
}
