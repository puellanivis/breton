package pes

import (
	"github.com/pkg/errors"
)

type packet struct {
	stream *Stream

	payload []byte
}

func (p *packet) preUnmarshal(b []byte) (int, error) {
	if b[0] != 0 || b[1] != 0 || b[2] != 1 {
		return 0, errors.Errorf("bad start prefix [% 2X]", b[:3])
	}

	if p.stream == nil {
		p.stream = new(Stream)
	}

	p.stream.ID = b[3]
	return (int(b[4]) << 8) | int(b[5]), nil
}

func (p *packet) unmarshal(b []byte) error {
	switch p.stream.ID {
	case idPaddingStream, idPrivateStream2:
		// Optional PES Header not present for these streams.

	default:
		l, err := p.stream.unmarshalHeader(b)
		if err != nil {
			return err
		}

		b = b[l:]
	}

	p.payload = append([]byte{}, b...)

	return nil
}

func (p *packet) Unmarshal(b []byte) error {
	l, err := p.preUnmarshal(b)
	if err != nil {
		return err
	}

	b = b[6 : 6+l] // enforce proper boundaries

	return p.unmarshal(b)
}

func (p *packet) Marshal() ([]byte, error) {
	var h []byte

	switch p.stream.ID {
	case idPaddingStream, idPrivateStream2:
		// Optional PES Header not present for these streams.

	default:
		var err error

		h, err = p.stream.marshalHeader()
		if err != nil {
			return nil, err
		}
	}

	l := len(h) + len(p.payload)

	if l > 0xffff {
		return nil, errors.Errorf("packet size too big: header:%d payload:%d", len(h), len(p.payload))
	}

	// len(Manditory PES Header) == 6
	out := make([]byte, 6+l)

	out[0] = 0
	out[1] = 0
	out[2] = 1
	out[3] = p.stream.ID

	out[4] = byte((l >> 8) & 0xff)
	out[5] = byte(l & 0xff)

	start := 6 + copy(out[6:], h)

	copy(out[start:], p.payload)

	return out, nil
}
