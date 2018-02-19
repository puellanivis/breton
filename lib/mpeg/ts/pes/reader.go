package pes

import (
	"bytes"
	"io"

	"github.com/pkg/errors"
)

type Reader struct {
	Stream

	buf bytes.Buffer
	src io.Reader
}

func NewReader(rd io.Reader) *Reader {
	return &Reader{
		src: rd,
	}
}

func (r *Reader) mustRead(b []byte) error {
	_, err := io.ReadFull(r.src, b)
	if err != io.EOF {
		return errors.WithStack(err)
	}

	return err
}

func (r *Reader) Read(b []byte) (n int, err error) {
	if r.buf.Len() > 0 {
		var err error

		n, err = r.buf.Read(b)
		if err != nil {
			return n, err
		}

		if n == len(b) {
			return n, nil
		}
	}

	hdr := make([]byte, 6)
	if err := r.mustRead(hdr); err != nil {
		if err == io.EOF && n != 0 {
			// only return err == io.EOF,
			// if we have not read anything from the buffer.
			return n, nil
		}

		return n, err
	}

	pkt := &packet{
		stream: &r.Stream,
	}

	l, err := pkt.preUnmarshal(hdr)
	if err != nil {
		return n, err
	}

	body := make([]byte, l)
	if err = r.mustRead(body); err != nil {
		return n, err
	}

	if err = pkt.unmarshal(body); err != nil {
		return n, err
	}

	m := copy(b[n:], pkt.payload)

	if m < len(pkt.payload) {
		_, err = r.buf.Write(pkt.payload[m:])
	}

	return n+m, err
}
