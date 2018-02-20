package dvb

import (
	"fmt"

	"github.com/pkg/errors"
	desc "github.com/puellanivis/breton/lib/mpeg/ts/descriptor"
)

type DVBServiceType uint8

const (
	DVBServiceTypeTV DVBServiceType = iota + 1
	DVBServiceTypeRadio
	DVBServiceTypeTeletext
	DVBServiceTypeHDTV     = 0x11
	DVBServiceTypeH264SDTV = 0x16
	DVBServiceTypeH264HDTV = 0x19
	DVBServiceTypeHEVCTV   = 0x1F
)

var dvbServiceTypeName = map[DVBServiceType]string{
	DVBServiceTypeTV:       "DVB-TV",
	DVBServiceTypeRadio:    "DVB-Radio",
	DVBServiceTypeTeletext: "DVB-Teletext",
	DVBServiceTypeHDTV:     "DVB-HDTV",
	DVBServiceTypeH264SDTV: "DVB-H.264-SDTV",
	DVBServiceTypeH264HDTV: "DVB-H.264-HDTV",
	DVBServiceTypeHEVCTV:   "DVB-HEVC-TV",
}

func (t DVBServiceType) String() string {
	if s, ok := dvbServiceTypeName[t]; ok {
		return s
	}

	return fmt.Sprintf("x%02X", uint8(t))
}

type DVBService struct {
	Type DVBServiceType

	Provider string
	Name     string
}

func (d *DVBService) String() string {
	return fmt.Sprintf("{DVBService %v P:%s N:%s}", d.Type, d.Provider, d.Name)
}

const (
	tagDVBService uint8 = 0x48
)

func init() {
	desc.Register(tagDVBService, func() desc.Descriptor { return new(DVBService) })
}

func (d *DVBService) Tag() uint8 {
	return tagDVBService
}

func (d *DVBService) Len() int {
	return 5 + len(d.Provider) + len(d.Name)
}

func (d *DVBService) Unmarshal(b []byte) error {
	if b[0] != tagDVBService {
		return errors.Errorf("TableID mismatch: x%02X != x%02X", b[0], tagDVBService)
	}

	l := int(b[1])

	b = b[2:]
	if len(b) < l {
		return errors.Errorf("unexpected end of byte-slice: %d < %d", len(b), l)
	}

	d.Type = DVBServiceType(b[0])
	b = b[1:]

	n := int(b[0])
	d.Provider = string(b[1 : 1+n])

	b = b[1+n:]

	n = int(b[0])
	d.Name = string(b[1 : 1+n])

	return nil
}

func (d *DVBService) Marshal() ([]byte, error) {
	l := 3 + len(d.Provider) + len(d.Name)
	if l > 0xFF {
		return nil, errors.Errorf("descriptor data field too large: %d", l)
	}

	b := make([]byte, d.Len())

	b[0] = tagDVBService
	b[1] = byte(l)
	b[2] = byte(d.Type)

	b = b[3:]

	b[0] = byte(len(d.Provider))
	n := copy(b[1:], d.Provider)

	b = b[1+n:]

	b[0] = byte(len(d.Name))
	copy(b[1:], d.Name)

	return b, nil
}
