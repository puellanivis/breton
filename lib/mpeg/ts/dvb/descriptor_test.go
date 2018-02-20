package dvb

import (
	"reflect"
	"testing"

	desc "github.com/puellanivis/breton/lib/mpeg/ts/descriptor"
)

var testServiceDescriptor = []byte{
	0x48, // descriptor_tag(service_descriptor_tag)
	18,   // descriptor_length
	0x01, // service_type(DVB-TV)

	6, 'F', 'F', 'm', 'p', 'e', 'g',                // service_provider_name
	9, 'S', 'e', 'r', 'v', 'i', 'c', 'e', '0', '1', // service_name
}

func TestServiceDescriptor(t *testing.T) {
	const (
		provider = "FFmpeg"
		name     = "Service01"
	)

	b := testServiceDescriptor

	d, err := desc.Unmarshal(b)
	if err != nil {
		t.Fatalf("Unmarshal: %+v", err)
	}

	service, ok := d.(*ServiceDescriptor)
	if !ok {
		t.Fatalf("wrong type, expected *ServiceDescriptor, got %T", d)
	}

	if service.Type != ServiceTypeTV {
		t.Errorf("wrong service type, expected x%02X got x%02X", uint8(ServiceTypeTV), uint8(service.Type))
	}

	if service.Provider != provider {
		t.Errorf("wrong provider name, expected %q got %q", provider, service.Provider)
	}

	if service.Name != name {
		t.Errorf("wrong name name, expected %q got %q", name, service.Name)
	}

	b, err = service.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %+v", err)
	}

	if !reflect.DeepEqual(b, testServiceDescriptor) {
		t.Errorf("Marshal: unexpected results\nexpected: [% 2X]\ngot:      [% 2X]", testServiceDescriptor, b)
	}
}
