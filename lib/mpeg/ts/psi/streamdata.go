package psi

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	desc "github.com/puellanivis/breton/lib/mpeg/ts/descriptor"
)

type StreamData struct {
	Type byte
	PID  uint16

	Descriptors []desc.Descriptor
}

func (esd *StreamData) String() string {
	out := []string{
		fmt.Sprintf("Type:x%02X", esd.Type),
		fmt.Sprintf("PID:x%04X", esd.PID),
	}

	for _, d := range esd.Descriptors {
		out = append(out, fmt.Sprintf("Desc:%v", d))
	}

	return fmt.Sprintf("{%s}", strings.Join(out, " "))
}

func (esd *StreamData) unmarshal(b []byte) (int, error) {
	esd.Type = b[0]
	esd.PID = uint16(b[1]&0x1f)<<8 | uint16(b[2])

	l := int(b[3]&0x3)<<8 | int(b[4])

	start := 5
	end := start + l

	for start < end {
		d, err := desc.Unmarshal(b[start:])
		if err != nil {
			return start, err
		}

		esd.Descriptors = append(esd.Descriptors, d)

		start += d.Len()
	}

	return start, nil
}

func (esd *StreamData) marshal() ([]byte, error) {
	data := make([]byte, 5)

	data[0] = esd.Type
	data[1] = byte((esd.PID >> 8) & 0x1f)
	data[2] = byte(esd.PID & 0xff)

	l := len(esd.Descriptors)
	if l > 0x3FF {
		return nil, errors.Errorf("too many descriptors: %d > 0x3FF ", l)
	}

	data[3] = byte((l >> 8) & 0x03)
	data[4] = byte(l & 0xFF)

	for _, d := range esd.Descriptors {
		b, err := d.Marshal()
		if err != nil {
			return nil, err
		}

		data = append(data, b...)
	}

	return data, nil
}
