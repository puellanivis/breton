package mapreduce

import (
	"fmt"
)

// A Range is a mathematical range defined as starting at Start and ending just before End.
// In mathematical notation [Start,End).
type Range struct {
	Start, End int
}

// String returns the Range in mathematical range notation.
func (r Range) String() string {
	return fmt.Sprintf("[%d,%d)", r.Start, r.End)
}

// Width returns the number of integers within the Range.
func (r Range) Width() int {
	return r.End - r.Start
}

// Add returns a new Range that is offset by the given amount, i.e. [Start+offset,End+offset).
func (r Range) Add(offset int) Range {
	return Range{
		Start: r.Start + offset,
		End:   r.End + offset,
	}
}
