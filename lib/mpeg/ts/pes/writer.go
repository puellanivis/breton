package pes

import (
	"io"
)

type Writer struct {
	Stream

	dst io.Writer
}

// NewWriter returns a Writer that encodes an Elementary Stream into a Packetized Elementary Stream,
// with the given Stream ID.
// Any furter Header values should be set before any Write is made.
func NewWriter(streamid byte, wr io.Writer) *Writer {
	return &Writer{
		Stream: Stream{
			ID: streamid,
		},

		dst: wr,
	}
}

// Write implements io.Writer.
func (w *Writer) Write(b []byte) (n int, err error) {
	pkt := &packet{
		stream:  &w.Stream,
		payload: b,
	}

	b2, err := pkt.Marshal()
	if err != nil {
		return 0, err
	}

	n, err = w.dst.Write(b2)
	if err != nil {
		return 0, err
	}

	if n < len(b2) {
		return 0, io.ErrShortWrite
	}

	return len(b), nil
}
