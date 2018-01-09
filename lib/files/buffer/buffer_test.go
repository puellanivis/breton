package buffer

import (
	"bytes"
	"testing"
)

func TestBuffer(t *testing.T) {
	b := new(Buffer)

	b.WriteString("ohai!")

	serialized := new(bytes.Buffer)
	_, err := b.WriteTo(serialized)
	if err != nil {
		t.Fatal("something went wrong writing to the serialized buffer", err)
	}

	if serialized.String() != "ohai!" {
		t.Fatal("a simple WriteString didn’t work", serialized)
	}
}
