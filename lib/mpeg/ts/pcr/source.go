package pcr

import (
	"time"
)

type Source struct{
	t time.Time
}

func NewSource() *Source {
	return &Source{
		t: time.Now(),
	}
}

func (s *Source) Read() *PCR {
	d := time.Since(s.t)

	return &PCR{
		base: uint64((d * 27) / time.Microsecond) & pcrModulo,
	}
}
