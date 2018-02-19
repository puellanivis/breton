package pes

import (
	"github.com/pkg/errors"
)

// Stream is a structure defining properties of a Primitive Elementary Stream.
type Stream struct {
	ID byte // Stream ID

	Header Header // Optional PES Header fields.
}

// Header is the Optional PES Header defined in SO/IEC 13818-1 and ITU-T H.222.0.
// It does not currently support any of the options that yield a variable length Header.
type Header struct {
	ScrambleControl byte

	Priority      bool
	DataAlignment bool
	Copyright     bool
	IsOriginal    bool

	padding []byte
}

func (h *Header) len() int {
	return 3 + len(h.padding)
}

const (
	markerBits = 0x80

	maskScramble  = 0x30
	shiftScramble = 4

	flagPriority  = 0x08
	flagAlignment = 0x04
	flagCopyright = 0x02
	flagOriginal  = 0x01
)

// Unmarshal fills in the values of an Optional PES Header from those encoded in the given byte-slice.
func (h *Header) Unmarshal(b []byte) error {
	length := 3 + int(b[2]) // full header length
	b = b[:length]          // enforce header length with slice boundaries

	h.ScrambleControl = (b[0] & maskScramble) >> shiftScramble
	h.Priority = b[0]&flagPriority != 0
	h.DataAlignment = b[0]&flagAlignment != 0
	h.Copyright = b[0]&flagCopyright != 0
	h.IsOriginal = b[0]&flagOriginal != 0

	// where the padding starts
	padStart := 3

	// we ignore all b[1] flags right now…
	// if we were to read one of them, then padStart += lengthOf(field)

	// we treat all of the rest of the header as “padding” for now
	h.padding = append([]byte{}, b[padStart:]...)

	return nil
}

// Marshal returns a byte-slice that is the encoding of a given Optional PES Header.
func (h *Header) Marshal() ([]byte, error) {
	if h.ScrambleControl&^0x03 != 0 {
		return nil, errors.Errorf("invalid scramble control: 0x%02x", h.ScrambleControl)
	}

	out := make([]byte, 3)

	out[0] = markerBits | (h.ScrambleControl << shiftScramble)

	if h.Priority {
		out[0] |= flagPriority
	}

	if h.DataAlignment {
		out[0] |= flagAlignment
	}

	if h.Copyright {
		out[0] |= flagCopyright
	}

	if h.IsOriginal {
		out[0] |= flagOriginal
	}

	out[1] = 0
	// The following fields are not supported and given is their presumed values:
	// They would need to be implemented in the following order, as they are concatted one-after-another.
	// PTS/DTS = both not included
	// ESCR = false
	// ES Rate = false
	// DSM Trick Mode = false
	// Additional Copy Info = false
	// CRC = false
	// Extension = false

	// Stuff any “padding” here at the end.
	out = append(out, h.padding...)

	// remaining header length
	out[2] = byte(len(out) - 3)

	return out, nil
}
