package dvb

import (
	"testing"

	desc "github.com/puellanivis/breton/lib/mpeg/ts/descriptor"
)

func TestDVBServe(t *testing.T) {
	const (
		provider = "FFmpeg"
		name     = "Service01"
	)

	b := []byte{
		0x48, 0, 0x01,
	}

	b = append(b, byte(len(provider)))
	b = append(b, provider...)

	b = append(b, byte(len(name)))
	b = append(b, name...)

	b[1] = byte(len(b) - 2)

	d, err := desc.Unmarshal(b)
	if err != nil {
		t.Fatal(err)
	}

	service, ok := d.(*DVBService)
	if !ok {
		t.Fatalf("wrong type, expected *DVBService, got %T", d)
	}

	if service.Type != DVBServiceTypeTV {
		t.Errorf("wrong service type, expected x%02X got x%02X", uint8(DVBServiceTypeTV), uint8(service.Type))
	}

	if service.Provider != provider {
		t.Errorf("wrong provider name, expected %q got %q", provider, service.Provider)
	}

	if service.Name != name {
		t.Errorf("wrong name name, expected %q got %q", name, service.Name)
	}

}
