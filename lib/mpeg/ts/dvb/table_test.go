package dvb

import (
	"reflect"
	"testing"

	desc "github.com/puellanivis/breton/lib/mpeg/ts/descriptor"
	"github.com/puellanivis/breton/lib/mpeg/ts/psi"
)

var testSDT = []byte{
	0x00,     // pointer_value
	0x42,     // table_id
	0xC0, 37, // section_syntax_indicator | private | section_length_bits
	0x00, 0x01, // transport_stream_id
	0xC1,       // reserved | version_number | current_indicator
	0x00, 0x00, // section_number, last_section_number
	0xff, 0x01, // original_network_id
	0xff, // reserved

	0x00, 0x01, // service_id = x0001
	0xFC,     // reserved_future_use
	0x80, 20, // running_status(0x4), descriptors_loop_length & 0xFF
	0x48, // descriptor_tag(service_descriptor_tag)
	18,   // descriptor_length
	0x01, // service_type(DVB-TV)

	6, 'F', 'F', 'm', 'p', 'e', 'g', // service_provider_name
	9, 'S', 'e', 'r', 'v', 'i', 'c', 'e', '0', '1', // service_name

	0xde, 0xad, 0xbe, 0xef, // fake CRC32
}

func TestSDT(t *testing.T) {
	b := testSDT

	p, err := psi.Unmarshal(b)
	if err != nil {
		t.Fatalf("Unmarshal: %+v", err)
	}

	sdt, ok := p.(*ServiceDescriptorTable)
	if !ok {
		t.Fatalf("wrong type, expected *ServiceDescriptorTable, got %T", p)
	}

	expected := &ServiceDescriptorTable{
		Syntax: &psi.SectionSyntax{
			TableIDExtension: 1,
			Current:          true,
		},
		OriginalNetworkID: 0xff01,
		Services: []*SDTService{
			&SDTService{
				ServiceID:     1,
				RunningStatus: Running,
				Descriptors: []desc.Descriptor{
					&ServiceDescriptor{
						Type:     ServiceTypeTV,
						Provider: "FFmpeg",
						Name:     "Service01",
					},
				},
			},
		},
		crc: 0xdeadbeef,
	}

	if !reflect.DeepEqual(sdt, expected) {
		t.Fatalf("Unmarshal: expected %v, got %v", expected, sdt)
	}

	b, err = expected.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %+v", err)
	}

	if !reflect.DeepEqual(b, testSDT) {
		t.Errorf("Marshal: unexpected results\nexpected: [% 2X]\ngot:      [% 2X]", testSDT, b)
	}
}
