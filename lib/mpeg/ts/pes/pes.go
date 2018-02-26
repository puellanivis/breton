package pes

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

const (
	idPaddingStream  = 0xBE
	idPrivateStream2 = 0xBF
)

// Stream is a structure defining properties of a Primitive Elementary Stream.
type Stream struct {
	ID byte // Stream ID

	Header // Optional PES Header fields.
}

// String implements fmt.Stringer
func (s *Stream) String() string {
	out := []string{
		fmt.Sprintf("ID:x%02X", s.ID),
	}

	if s.ID != idPaddingStream && s.ID != idPrivateStream2 {
		if s.ScrambleControl != 0 {
			out = append(out, fmt.Sprintf("Scramble:x%X", s.ScrambleControl))
		}

		if s.Priority {
			out = append(out, "PRI")
		}

		if s.DataAlignment {
			out = append(out, "ALIGN")
		}

		if s.Copyright {
			out = append(out, "COPYRIGHT")
		}

		if s.IsOriginal {
			out = append(out, "ORIG")
		}

		if s.PTS != nil {
			out = append(out, fmt.Sprintf("PTS:x%09X", *s.PTS))
		}

		if s.DTS != nil {
			out = append(out, fmt.Sprintf("DTS:x%09X", *s.DTS))
		}

		if s.extFlags != 0 {
			out = append(out, fmt.Sprintf("flags:%02X", s.extFlags))
		}

		if len(s.padding) > 0 {
			out = append(out, fmt.Sprintf("padding[%d]{% 2X}", len(s.padding), s.padding))
		}
	}

	return fmt.Sprintf("{%s}", strings.Join(out, " "))
}

// Header is the Optional PES Header defined in SO/IEC 13818-1 and ITU-T H.222.0.
// It does not currently support any of the options that yield a variable length Header.
type Header struct {
	ScrambleControl byte

	Priority      bool
	DataAlignment bool
	Copyright     bool
	IsOriginal    bool

	PTS *uint64
	DTS *uint64

	extLen int

	extFlags byte
	padding  []byte
}

const (
	markerBits = 0x80

	maskScramble  = 0x30
	shiftScramble = 4

	flagPriority  = 0x08
	flagAlignment = 0x04
	flagCopyright = 0x02
	flagOriginal  = 0x01

	flagPTSDTS = 0xC0
)

// unmarshalHeader fills in the values of an Optional PES Header from those encoded in the given byte-slice.
func (h *Header) unmarshalHeader(b []byte) (int, error) {
	length := 3 + int(b[2]) // full header length
	b = b[:length]          // enforce header length with slice boundaries

	h.ScrambleControl = (b[0] & maskScramble) >> shiftScramble
	h.Priority = b[0]&flagPriority != 0
	h.DataAlignment = b[0]&flagAlignment != 0
	h.Copyright = b[0]&flagCopyright != 0
	h.IsOriginal = b[0]&flagOriginal != 0

	// where the padding starts
	padStart := 3

	switch b[1] & flagPTSDTS {
	case 0x40:
		return length, errors.New("invalid PTS/DTS flag value")

	case 0x80:
		h.PTS = decodeTS(b[padStart:])
		padStart += 5

	case 0xC0:
		h.PTS = decodeTS(b[padStart:])
		padStart += 5

		h.DTS = decodeTS(b[padStart:])
		padStart += 5
	}

	// we ignore all b[1] flags right now…
	// if we were to read one of them, then padStart += lengthOf(field)
	h.extFlags = b[1] &^ 0xc0

	// we treat all of the rest of the header as “padding” for now
	h.padding = append([]byte{}, b[padStart:]...)

	return length, nil
}

// marshalHeader returns a byte-slice that is the encoding of a given Optional PES Header.
func (h *Header) marshalHeader() ([]byte, error) {
	if h.ScrambleControl&^0x03 != 0 {
		return nil, errors.Errorf("invalid scramble control: 0x%02X", h.ScrambleControl)
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

	var pts, dts []byte

	if h.PTS != nil {
		pts = encodeTS(*h.PTS)
		pts[0] |= 0x20
		out[1] |= 0x80
	}

	if h.DTS != nil {
		dts = encodeTS(*h.DTS)
		pts[0] |= 0x10
		dts[0] |= 0x10
		out[1] |= 0x40
	}

	if len(pts) > 0 {
		out = append(out, pts...)
	}

	if len(dts) > 0 {
		out = append(out, dts...)
	}

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
