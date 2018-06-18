package mapreduce

import (
	"fmt"
)

type Range struct {
	Start, End int
}

func (r Range) String() string {
	return fmt.Sprintf("[%d,%d)", r.Start, r.End)
}

func (r Range) Width() int {
	return r.End - r.Start
}

func (r Range) Add(off int) Range {
	return Range{
		Start: r.Start + off,
		End:   r.End + off,
	}
}
