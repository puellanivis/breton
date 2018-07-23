package packet

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// TransportScrambleControl is an enum of what kind of MPEG Transport Scramble Control is to be used.
type TransportScrambleControl byte

// TransportScrambleControl enum values.
const (
	ScrambleNone TransportScrambleControl = iota
	scrambleReserve
	ScrambleEven
	ScrambleOdd
)

// Packet defines a single MPEG-TS packet.
type Packet struct {
	PID      uint16
	TEI      bool
	PUSI     bool
	Priority bool

	ScrambleControl TransportScrambleControl
	*AdaptationField
	Continuity byte

	Payload []byte
}

func (p *Packet) String() string {
	var out []string

	out = append(out, fmt.Sprintf("PID:x%04X", p.PID), fmt.Sprintf("[%X]", p.Continuity))
	if p.TEI {
		out = append(out, "TEI")
	}
	if p.PUSI {
		out = append(out, "PUSI")
	}
	if p.Priority {
		out = append(out, "PRI")
	}

	switch p.ScrambleControl {
	case ScrambleNone:
	case ScrambleEven:
		out = append(out, "EVEN")
	case ScrambleOdd:
		out = append(out, "ODD")
	}

	if p.AdaptationField != nil {
		out = append(out, fmt.Sprintf("AF:%+v", p.AdaptationField))
	}

	if len(p.Payload) > 0 {
		pl := fmt.Sprintf("Payload[%d]", len(p.Payload))

		if len(p.Payload) > 16 {
			pl = fmt.Sprintf("%s{ % 2X… }", pl, p.Payload[:16])
		} else {
			pl = fmt.Sprintf("%s{ % 2X }", pl, p.Payload)
		}

		out = append(out, pl)
	}

	return fmt.Sprintf("{%s}", strings.Join(out, " "))
}

const (
	flagTEI      = 0x80
	flagPUSI     = 0x40
	flagPriority = 0x20

	flagPayload         = 0x10
	flagAdaptationField = 0x20

	// Length is how long an MPEG-TS packet is.
	Length = 188

	// HeaderLength is the length of an MPEG-TS packet.
	HeaderLength = 4

	// MaxPayload is the maximum length of payload that may be put into a packet.
	MaxPayload = Length - HeaderLength
)

// Bytes returns the payload of the Packet as a byte slice.
func (p *Packet) Bytes() []byte {
	return p.Payload
}

// Unmarshal decodes a byte slice into the Packet.
func (p *Packet) Unmarshal(b []byte) error {
	if len(b) != Length || b[0] != 'G' {
		return errors.Errorf("invalid packet %v", b[:4])
	}

	p.TEI = (b[1] & flagTEI) != 0
	p.PUSI = (b[1] & flagPUSI) != 0
	p.Priority = (b[1] & flagPriority) != 0

	p.PID = (uint16(b[1]&0x1f) << 8) | uint16(b[2])

	p.ScrambleControl = TransportScrambleControl((b[3] >> 6) & 0x03)
	p.Continuity = b[3] & 0x0f

	start := 4

	if b[3]&flagAdaptationField != 0 {
		af := new(AdaptationField)

		l, err := af.unmarshal(b[start:])
		if err != nil {
			return err
		}

		p.AdaptationField = af

		start += l
	}

	if b[3]&flagPayload != 0 {
		p.Payload = append([]byte{}, b[start:]...)
	}

	return nil
}

var fullPadding = bytes.Repeat([]byte{0xFF}, Length)

// Marshal encodes a Packet into a byte slice.
func (p *Packet) Marshal() ([]byte, error) {
	if p.PID > 0x1fff {
		return nil, errors.Errorf("PID %d is greater than maximum 0x1fff", p.PID)
	}

	packet := make([]byte, Length)

	packet[0] = 'G'

	if p.TEI {
		packet[1] |= flagTEI
	}

	if p.PUSI {
		packet[1] |= flagPUSI
	}

	if p.Priority {
		packet[1] |= flagPriority
	}

	packet[1] |= byte((p.PID >> 8) & 0x1f)
	packet[2] = byte(p.PID & 0xff)

	packet[3] = byte((p.ScrambleControl&0x03)<<6) | byte(p.Continuity&0x0f)

	start := 4
	if p.AdaptationField != nil {
		packet[3] |= flagAdaptationField

		b, err := p.AdaptationField.marshal()
		if err != nil {
			return nil, err
		}

		n := copy(packet[start:], b)
		start += n
	}

	if len(p.Payload) > 0 {
		packet[3] |= flagPayload

		n := copy(packet[start:], p.Payload)
		start += n

		if n < len(p.Payload) {
			return nil, errors.Errorf("short packet: %d < %d", n, len(p.Payload))
		}
	}

	copy(packet[start:], fullPadding)

	return packet, nil
}
