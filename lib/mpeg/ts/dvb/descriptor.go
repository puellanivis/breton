package dvb

import (
	"fmt"

	"github.com/pkg/errors"
	desc "github.com/puellanivis/breton/lib/mpeg/ts/descriptor"
)

// ServiceType is an enum encoding DVB standards indicating service types for MPEG-TS.
type ServiceType uint8

// SesrviceType enum values.
const (
	ServiceTypeTV ServiceType = iota + 1
	ServiceTypeRadio
	ServiceTypeTeletext
	ServiceTypeHDTV     = 0x11
	ServiceTypeH264SDTV = 0x16
	ServiceTypeH264HDTV = 0x19
	ServiceTypeHEVCTV   = 0x1F
)

var dvbServiceTypeNames = map[ServiceType]string{
	ServiceTypeTV:       "TV",
	ServiceTypeRadio:    "Radio",
	ServiceTypeTeletext: "Teletext",
	ServiceTypeHDTV:     "HDTV",
	ServiceTypeH264SDTV: "H.264-SDTV",
	ServiceTypeH264HDTV: "H.264-HDTV",
	ServiceTypeHEVCTV:   "HEVC-TV",
}

func (t ServiceType) String() string {
	if s, ok := dvbServiceTypeNames[t]; ok {
		return s
	}

	return fmt.Sprintf("x%02X", uint8(t))
}

// ServiceDescriptor defines a DVB Service Descriptor for the MPEG-TS standard.
type ServiceDescriptor struct {
	Type ServiceType

	Provider string
	Name     string
}

func (d *ServiceDescriptor) String() string {
	return fmt.Sprintf("{DVB:ServiceDesc %v P:%q N:%q}", d.Type, d.Provider, d.Name)
}

const (
	tagDVBService uint8 = 0x48
)

func init() {
	desc.Register(tagDVBService, func() desc.Descriptor { return new(ServiceDescriptor) })
}

// Tag implements mpeg/ts/descriptor.Descriptor.
func (d *ServiceDescriptor) Tag() uint8 {
	return tagDVBService
}

// Len returns the length in bytes that the ServiceDescriptor would encode into, and implements mpeg/ts/descriptor.Descriptor.
func (d *ServiceDescriptor) Len() int {
	return 5 + len(d.Provider) + len(d.Name)
}

// Unmarshal decodes a byte slice into the ServiceDescriptor.
func (d *ServiceDescriptor) Unmarshal(b []byte) error {
	if b[0] != tagDVBService {
		return errors.Errorf("descriptor_tag mismatch: x%02X != x%02X", b[0], tagDVBService)
	}

	l := int(b[1])

	b = b[2:]
	if len(b) < l {
		return errors.Errorf("unexpected end of byte-slice: %d < %d", len(b), l)
	}

	d.Type = ServiceType(b[0])
	b = b[1:]

	n := int(b[0])
	d.Provider = string(b[1 : 1+n])

	b = b[1+n:]

	n = int(b[0])
	d.Name = string(b[1 : 1+n])

	return nil
}

// Marshal encodes the ServiceDescriptor into a byte slice.
func (d *ServiceDescriptor) Marshal() ([]byte, error) {
	l := 3 + len(d.Provider) + len(d.Name)
	if l > 0xFF {
		return nil, errors.Errorf("descriptor data field too large: %d", l)
	}

	b := make([]byte, d.Len())

	b[0] = tagDVBService
	b[1] = byte(l)
	b[2] = byte(d.Type)

	b[3] = byte(len(d.Provider))
	n := copy(b[4:], d.Provider)

	b[4+n] = byte(len(d.Name))
	copy(b[5+n:], d.Name)

	return b, nil
}
