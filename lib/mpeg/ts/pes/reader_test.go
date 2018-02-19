package pes

import (
	"bytes"
	"reflect"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	expected := [][]byte{
		[]byte("Hello World"),
		[]byte("foo bar"),
	}

	b := []byte{
		0, 0, 1,
		0x42,
		0, byte(3 + len(expected[0])),
		0x80, 0x00, 0x00,
	}
	b = append(b, expected[0]...)

	b = append(b, []byte{
		0, 0, 1,
		0x42,
		0, byte(3 + len(expected[1])),
		0x80, 0x00, 0x00,
	}...)
	b = append(b, expected[1]...)

	var all []byte
	for i := range expected {
		all = append(all, expected[i]...)
	}

	rd := &Reader{
		src: bytes.NewReader(b),
	}

	for i := range expected {
		output := make([]byte, len(all))
		n, err := rd.Read(output)
		if err != nil {
			t.Fatal(err)
		}

		if n != len(expected[i]) {
			t.Errorf("expected to read %d bytes, but read %d bytes", len(expected[i]), n)
		}

		output = output[:n]

		if !reflect.DeepEqual(output, expected[i]) {
			t.Errorf("expected to read %v, but read %v", expected[i], output)
		}
	}

	rd = &Reader{
		src: bytes.NewReader(b),
	}

	// test reading with a super small buffer.
	output := make([]byte, 3)
	for i := 0; i < len(all); i += 3 {
		expectedLen := 3
		if i+3 > len(all) {
			expectedLen = len(all) - i
		}

		n, err := rd.Read(output)
		if err != nil {
			t.Fatal(err)
		}

		if n != expectedLen {
			t.Errorf("expected to read %d bytes, but read %d bytes", expectedLen, n)
		}

		output := output[:n]
		expected := all[i : i+expectedLen]

		if !reflect.DeepEqual(output, expected) {
			t.Errorf("expected to read %v, but read %v", expected, output)
		}
	}
}
