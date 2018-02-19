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
	for r.buf.Len() >= len(b) {
		return r.buf.Read(b)
	}

	hdr := make([]byte, 6)
	if err := r.mustRead(hdr); err != nil {
		return 0, err
	}

	pkt := &packet{
		stream: &r.Stream,
	}

	l, err := pkt.preUnmarshal(hdr)
	if err != nil {
		return 0, err
	}

	body := make([]byte, l)
	if err := r.mustRead(body); err != nil {
		return 0, err
	}

	if err := pkt.unmarshal(body); err != nil {
		return 0, err
	}

	if _, err = r.buf.Write(pkt.payload); err != nil {
		return 0, err
	}

	return r.buf.Read(b)
}
