package psi

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	desc "github.com/puellanivis/breton/lib/mpeg/ts/descriptor"
)

type ElementaryStreamData struct {
	Type byte
	PID  uint16

	Descriptors []desc.Descriptor
}

func (esd *ElementaryStreamData) String() string {
	out := []string{
		fmt.Sprintf("Type:x%02x", esd.Type),
		fmt.Sprintf("PID:x%04x", esd.PID),
	}

	for _, d := range esd.Descriptors {
		out = append(out, fmt.Sprintf("Desc:%v", d))
	}

	return fmt.Sprintf("{%s}", strings.Join(out, " "))
}

func (esd *ElementaryStreamData) Unmarshal(b []byte) error {
	esd.Type = b[0]
	esd.PID = uint16(b[1]&0x1f)<<8 | uint16(b[2])

	l := int(b[3]&0x3)<<8 | int(b[4])

	start := 5

	esd.Descriptors = make([]desc.Descriptor, l)
	for i := range esd.Descriptors {
		d, err := desc.Unmarshal(b[start:])
		if err != nil {
			return err
		}

		esd.Descriptors[i] = d

		start += d.Len()
	}

	return nil
}

func (esd *ElementaryStreamData) Marshal() ([]byte, error) {
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

type PMT struct {
	Syntax *SectionSyntax

	PCRPID      uint16
	Descriptors []desc.Descriptor
	Streams     []*ElementaryStreamData

	crc uint32
}

func (pmt *PMT) String() string {
	out := []string{
		"PMT",
	}

	if pmt.Syntax != nil {
		out = append(out, fmt.Sprint(pmt.Syntax))
	}

	if pmt.PCRPID != 0x1fff {
		out = append(out, fmt.Sprintf("PCRPID:x%04x", pmt.PCRPID))
	}

	for _, d := range pmt.Descriptors {
		out = append(out, fmt.Sprintf("Desc:%v", d))
	}

	for _, esd := range pmt.Streams {
		out = append(out, fmt.Sprintf("Stream:%v", esd))
	}

	return fmt.Sprintf("{%s}", strings.Join(out, " "))
}

const (
	tableidPMT = 0x02
)

func init() {
	Register(tableidPMT, func() PSI { return new(PMT) })
}

func (pmt *PMT) TableID() uint8 {
	return tableidPMT
}

func (pmt *PMT) Unmarshal(b []byte) error {
	if b[0] != tableidPMT {
		return errors.Errorf("table_id mismatch: x%02X != x%02X", b[0], tableidPMT)
	}

	syn, data, crc, err := CommonUnmarshal(b)
	if err != nil {
		return err
	}

	pmt.Syntax = syn
	pmt.crc = crc

	pmt.PCRPID = uint16(data[0]&0x1F)<<8 | uint16(data[1])

	start := 4
	pinfo_length := int(data[2]&0x03)<<8 | int(data[3])

	pmt.Descriptors = make([]desc.Descriptor, pinfo_length)
	for i := range pmt.Descriptors {
		d, err := desc.Unmarshal(data[start:])
		if err != nil {
			return err
		}

		pmt.Descriptors[i] = d

		start += d.Len()
	}

	for start < len(data) {
		b := data[start:]

		esd := new(ElementaryStreamData)
		if err := esd.Unmarshal(b); err != nil {
			return err
		}

		pmt.Streams = append(pmt.Streams, esd)

		start += 5
		for _, d := range esd.Descriptors {
			start += d.Len()
		}
	}

	return nil
}

func (pmt *PMT) Marshal() ([]byte, error) {
	data := make([]byte, 4)

	data[0] = byte((pmt.PCRPID >> 8) & 0x1F)
	data[1] = byte(pmt.PCRPID & 0xFF)

	l := len(pmt.Descriptors)
	if l > 0x3FF {
		return nil, errors.Errorf("too many descriptors: %d > 0x3FF ", l)
	}

	data[2] = byte((l >> 8) & 0x03)
	data[3] = byte(l & 0xFF)

	for _, d := range pmt.Descriptors {
		b, err := d.Marshal()
		if err != nil {
			return nil, err
		}

		data = append(data, b...)
	}

	for _, s := range pmt.Streams {
		b, err := s.Marshal()
		if err != nil {
			return nil, err
		}

		data = append(data, b...)
	}

	return CommonMarshal(tableidPMT, false, pmt.Syntax, data)
}
