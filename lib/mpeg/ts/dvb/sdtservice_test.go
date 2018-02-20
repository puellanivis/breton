package dvb

import (
	"reflect"
	"testing"

	desc "github.com/puellanivis/breton/lib/mpeg/ts/descriptor"
)

var testSDTService = []byte{
	0x00, 0x01, // service_id = x0001
	0xfd,     // reserved_future_use(0xfc) | EIT_present_follow_flag
	0xf0, 20, // running_status(0x7) | free_CA_mode, descriptors_loop_length & 0xFF
	0x48, // descriptor_tag(service_descriptor_tag)
	18,   // descriptor_length
	0x01, // service_type(DVB-TV)

	6, 'F', 'F', 'm', 'p', 'e', 'g', // service_provider_name
	9, 'S', 'e', 'r', 'v', 'i', 'c', 'e', '0', '1', // service_name
}

func TestSDTService(t *testing.T) {
	b := testSDTService

	s := new(SDTService)

	l, err := s.Unmarshal(b)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	if l != len(testSDTService) {
		t.Errorf("SDTService.Unmarshal returned wrong length read, expected %d, got %d", len(testSDTService), l)
	}

	expected := &SDTService{
		ServiceID:     1,
		EITSchedule:   false,
		EITPresent:    true,
		RunningStatus: 0x7,
		FreeCA:        true,
		Descriptors: []desc.Descriptor{
			&ServiceDescriptor{
				Type:     ServiceTypeTV,
				Provider: "FFmpeg",
				Name:     "Service01",
			},
		},
	}

	if !reflect.DeepEqual(s, expected) {
		t.Errorf("Unmarshal: expected %v, go %v", expected, s)
	}

	b, err = expected.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %+v", err)
	}

	if !reflect.DeepEqual(b, testSDTService) {
		t.Errorf("Marshal: unexpected results\nexpected: [% 2X]\ngot:      [% 2X]", testSDTService, b)
	}
}
