package ts

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/puellanivis/breton/lib/mpeg/ts/psi"
)

type TransportScrambleControl byte

const (
	ScrambleNone TransportScrambleControl = iota
	scrambleReserve
	ScrambleEven
	ScrambleOdd
)

type Packet struct {
	PID      uint16
	TEI      bool
	PUSI     bool
	Priority bool

	TransportScrambleControl
	*AdaptationField
	Continuity byte

	payload []byte
	PSI     psi.PSI
}

func (p *Packet) String() string {
	var out []string

	out = append(out, fmt.Sprintf("PID:x%06X", p.PID), fmt.Sprintf("[%X]", p.Continuity))
	if p.TEI {
		out = append(out, "TEI")
	}
	if p.PUSI {
		out = append(out, "PUSI")
	}
	if p.Priority {
		out = append(out, "PRI")
	}

	switch p.TransportScrambleControl {
	case ScrambleNone:
	case ScrambleEven:
		out = append(out, "EVEN")
	case ScrambleOdd:
		out = append(out, "ODD")
	}

	if p.AdaptationField != nil {
		out = append(out, fmt.Sprintf("AF:%+v", p.AdaptationField))
	}

	switch {
	case p.PSI != nil:
		out = append(out, fmt.Sprintf("%v", p.PSI))

	case len(p.payload) > 0:
		pl := fmt.Sprintf("payload[%d]", len(p.payload))

		if len(p.payload) > 4 {
			pl = fmt.Sprintf("%s{ % 2x… }", pl, p.payload[:4])
		}

		out = append(out, pl)
	}

	return fmt.Sprintf("{%s}", strings.Join(out, " "))
}

const (
	flagTEI      = 0x80
	flagPUSI     = 0x40
	flagPriority = 0x20

	maskTSC        = 0xc0
	maskAFC        = 0x30
	maskContinuity = 0xf

	pktLen = 188
)

func (p *Packet) Bytes() []byte {
	return p.payload
}

func (p *Packet) Unmarshal(b []byte) error {
	if len(b) != pktLen || b[0] != 'G' {
		return errors.Errorf("invalid packet %v", b[:4])
	}

	p.TEI = (b[1] & flagTEI) != 0
	p.PUSI = (b[1] & flagPUSI) != 0
	p.Priority = (b[1] & flagPriority) != 0

	p.PID = (uint16(b[1]&0x1f) << 8) | uint16(b[2])

	p.TransportScrambleControl = TransportScrambleControl((b[3] >> 6) & 0x03)
	p.Continuity = b[3] & 0x0f

	start := 4

	if b[3]&0x20 != 0 {
		af := new(AdaptationField)

		l, err := af.unmarshal(b[start:])
		if err != nil {
			return err
		}

		p.AdaptationField = af

		start += l
	}

	if b[3]&0x10 != 0 {
		b := b[start:]

		if p.PUSI {
			if b[0] != 0 || b[1] != 0 || b[2] != 1 {
				psi, err := psi.Unmarshal(b)
				if err != nil {
					return err
				}

				p.PSI = psi
				return nil
			}
		}

		p.payload = append([]byte{}, b...)
	}

	return nil
}

func (p *Packet) Marshal() ([]byte, error) {
	if p.PID > 0x1fff {
		return nil, errors.Errorf("PID %d is greater than maximum 0x1fff", p.PID)
	}

	packet := make([]byte, pktLen)

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

	packet[1] |= byte((p.PID & 0x1f) >> 8)
	packet[2] = byte(p.PID & 0xff)

	packet[3] = byte((p.TransportScrambleControl&0x03)<<6) | byte(p.Continuity&0x0f)

	start := 4
	if p.AdaptationField != nil {
		packet[3] |= 0x10

		b, err := p.AdaptationField.marshal()
		if err != nil {
			return nil, err
		}

		n := copy(packet[start:], b)
		start += n
	}

	switch {
	case p.PSI != nil:
		packet[3] |= 0x20

		b, err := p.PSI.Marshal()
		if err != nil {
			return nil, err
		}

		n := copy(packet[start:], b)
		start += n

	case len(p.payload) > 0:
		packet[3] |= 0x20

		n := copy(packet[start:], p.payload)
		start += n
	}

	return packet, nil
}
