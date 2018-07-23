// Package buffer (DO NOT USE) is a preliminary library futzing with how to efficiently hold mutable byte-slices.
package buffer

import (
	"errors"
	"io"
	"sync"
)

// Buffer is a DO NOT USE work-in-progress buffer to efficiently hold mutable byte-slices.
type Buffer struct {
	sync.RWMutex

	segments []*segment
}

// Len returns the entire length of the given Buffer.
func (b *Buffer) Len() int {
	b.RLock()
	defer b.RUnlock()

	var l int

	for _, s := range b.segments {
		l += s.Len()
	}

	return l
}

// WriteString writes a string into the Buffer.
func (b *Buffer) WriteString(s string) (n int, err error) {
	return b.Write([]byte(s))
}

// Write implements io.Writer.
func (b *Buffer) Write(buf []byte) (n int, err error) {
	b.Lock()
	defer b.Unlock()

	if len(b.segments) < 1 {
		b.segments = append(b.segments, new(segment))
	}

	s := b.segments[len(b.segments)-1]

	for {
		t := s.Append(buf)
		n += t

		if t < len(buf) {
			buf = buf[t:]

			s = new(segment)
			b.segments = append(b.segments, s)

			continue
		}

		return n, nil
	}
}

// ReadAt implements io.ReaderAt.
func (b *Buffer) ReadAt(buf []byte, off int64) (n int, err error) {
	if off != 0 {
		return 0, errors.New("Unsupported")
	}

	b.RLock()
	defer b.RUnlock()

	for _, s := range b.segments {
		if n >= len(buf) {
			return n, nil
		}

		n += copy(buf[n:], s.Bytes())
	}

	return n, nil
}

// WriteTo implements io.WriterTo.
func (b *Buffer) WriteTo(w io.Writer) (n int64, err error) {
	b.RLock()
	defer b.RUnlock()

	for _, s := range b.segments {
		t, err := w.Write(s.Bytes())
		n += int64(t)

		if err != nil {
			return n, err
		}
	}

	return n, nil
}
