package pes

import (
	"bytes"
	"reflect"
	"testing"
)

func TestMarshal(t *testing.T) {
	b := new(bytes.Buffer)

	wr := &Writer{
		dst: b,
		Stream: Stream{
			ID: 0x42,
		},
	}

	input := []byte("Hello World")

	expected := []byte{
		0, 0, 1,
		0x42,
		0, byte(3+len(input)),
		0x80, 0x00, 0x00,
	}
	expected = append(expected, input...)
	expectedWriteLength := len(expected)

	n, err := wr.Write(input)
	if err != nil {
		t.Fatal(err)
	}

	if n != expectedWriteLength {
		t.Errorf("expected to write %d bytes, but wrote %d", expectedWriteLength, n)
	}

	output := b.Bytes()
	if !reflect.DeepEqual(output, expected) {
		t.Errorf("expected to write %v, but wrote %v", expected, output)
	}

	// Here, we’re checking minimally that a second write does the same expected thing, but twice.
	expected = append(expected, expected...)

	n, err = wr.Write(input)
	if err != nil {
		t.Fatal(err)
	}

	if n != expectedWriteLength {
		t.Errorf("expected to write %d bytes, but wrote %d", expectedWriteLength, n)
	}

	output = b.Bytes()
	if !reflect.DeepEqual(output, expected) {
		t.Errorf("expected to write %v, but wrote %v", expected, output)
	}
}
