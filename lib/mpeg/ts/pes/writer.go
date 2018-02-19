package pes

import (
	"io"
)

type Writer struct {
	Stream

	dst io.Writer
}

func (w *Writer) Write(b []byte) (n int, err error) {
	pkt := &packet{
		stream: &w.Stream,
		payload: b,
	}

	b2, err := pkt.Marshal()
	if err != nil {
		return 0, err
	}

	return w.dst.Write(b2)
}
