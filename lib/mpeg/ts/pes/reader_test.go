package pes

import (
	"bytes"
	"reflect"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	expected := []byte("Hello World")

	b := []byte{
		0, 0, 1,
		0x42,
		0, byte(3+len(expected)),
		0x80, 0x00, 0x00,
	}
	b = append(b, expected...)

	rd := &Reader{
		src: bytes.NewReader(b),
	}

	output := make([]byte, len(expected))
	n, err := rd.Read(output)
	if err != nil {
		t.Fatal(err)
	}

	if n != len(output) {
		t.Errorf("expected to read %d bytes, but read %d bytes", len(output), n)
	}

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("expected to read %v, but read %v", expected, output)
	}
}
