package dvb

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/puellanivis/breton/lib/mpeg/ts/psi"
)

type ServiceDescriptorTable struct {
	Syntax *psi.SectionSyntax

	OriginalNetworkID uint16

	Services []*Service

	crc uint32
}

const (
	tableidSDT = 0x42
)

func init() {
	psi.Register(tableidSDT, func() psi.PSI { return new(ServiceDescriptorTable) })
}

func (sdt *ServiceDescriptorTable) TableID() uint8 {
	return tableidSDT
}

func (sdt *ServiceDescriptorTable) SectionSyntax() *psi.SectionSyntax {
	return sdt.Syntax
}

func (sdt *ServiceDescriptorTable) String() string {
	out := []string{
		"DVB:SDT",
	}

	if sdt.Syntax != nil {
		out = append(out, fmt.Sprint(sdt.Syntax))
	}

	out = append(out, fmt.Sprintf("OrigNetID:%04x", sdt.OriginalNetworkID))

	for _, s := range sdt.Services {
		out = append(out, s.String())
	}

	return fmt.Sprintf("{%s}", strings.Join(out, " "))
}

func (sdt *ServiceDescriptorTable) Unmarshal(b []byte) error {
	if b[0] != tableidSDT {
		return errors.Errorf("table_id mismatch: x%02X != x%02X", b[0], tableidSDT)
	}

	syn, data, crc, err := psi.CommonUnmarshal(b)
	if err != nil {
		return err
	}

	sdt.Syntax = syn
	sdt.crc = crc

	sdt.OriginalNetworkID = uint16(data[0])<<8 | uint16(data[1])

	start := 3 // original_network_id uint16 + reserved_future_use uint8
	for start < len(data) {
		s := new(Service)

		l, err := s.unmarshal(data[start:])
		if err != nil {
			return err
		}

		sdt.Services = append(sdt.Services, s)

		start += l
	}

	return nil
}

func (sdt *ServiceDescriptorTable) Marshal() ([]byte, error) {
	data := make([]byte, 3)

	data[0] = byte(sdt.OriginalNetworkID >> 8 & 0xFF)
	data[1] = byte(sdt.OriginalNetworkID & 0xFF)
	data[2] = 0xFF // reserved_future_use

	for _, s := range sdt.Services {
		sb, err := s.marshal()
		if err != nil {
			return nil, err
		}

		data = append(data, sb...)
	}

	return psi.CommonMarshal(tableidSDT, true, sdt.Syntax, data)
}
