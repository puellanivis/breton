package dvb

import (
	"reflect"
	"testing"

	desc "github.com/puellanivis/breton/lib/mpeg/ts/descriptor"
)

var testService = []byte{
	0x00, 0x01, // service_id = x0001
	0xfd,     // reserved_future_use(0xfc) | EIT_present_follow_flag
	0xf0, 20, // running_status(0x7) | free_CA_mode, descriptors_loop_length & 0xFF
	0x48, // descriptor_tag(service_descriptor_tag)
	18,   // descriptor_length
	0x01, // service_type(DVB-TV)

	6, 'F', 'F', 'm', 'p', 'e', 'g', // service_provider_name
	9, 'S', 'e', 'r', 'v', 'i', 'c', 'e', '0', '1', // service_name
}

func TestService(t *testing.T) {
	b := testService

	s := new(Service)

	l, err := s.unmarshal(b)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	if l != len(testService) {
		t.Errorf("Service.Unmarshal returned wrong length read, expected %d, got %d", len(testService), l)
	}

	expected := &Service{
		ID:            1,
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

	b, err = expected.marshal()
	if err != nil {
		t.Fatalf("Marshal: %+v", err)
	}

	if !reflect.DeepEqual(b, testService) {
		t.Errorf("Marshal: unexpected results\nexpected: [% 2X]\ngot:      [% 2X]", testService, b)
	}
}
