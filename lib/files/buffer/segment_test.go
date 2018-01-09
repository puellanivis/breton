package buffer

import (
	"bytes"
	"testing"
)

func TestSegment(t *testing.T) {
	s := new(segment)

	if s.Len() > 0 {
		t.Fatal("new(segment) length greater than 0")
	}

	n := s.Append(nil)
	if n != 0 {
		t.Fatal("segment.Append returned unexpected number of bytes appended")
	}

	n = s.Append([]byte{0x02, 0x03, 0x05})
	if n != 3 {
		t.Fatal("segment.Append returned unexpected number of bytes appended")
	}

	b := s.Bytes()
	if len(b) != 3 {
		t.Fatal("segment.Bytes didn’t return a length 3 slice instead, got ", len(b))
	}

	n = s.Append([]byte{0x07, 0x0B, 0x0D})
	if n != 3 {
		t.Fatal("segment.Append returned unexpected number of bytes appended")
	}

	b = s.Bytes()
	if len(b) != 6 {
		t.Fatal("segment.Bytes didn’t return a length 6 slice instead, got ", len(b))
	}

	if bytes.Compare(b, []byte{0x02, 0x03, 0x05, 0x07, 0x0B, 0x0D}) != 0 {
		t.Fatal("segment.Bytes doesn’t contain expected values: ", b)
	}

	s2 := s.Copy()
	b2 := s2.Bytes()

	if bytes.Compare(b2, b) != 0 {
		t.Fatal("segment.Copy didn’t return the same values", b2)
	}

	n = s2.Append([]byte{0x11})
	if n != 1 {
		t.Fatal("segment.Append returnned unexpected number of bytes appended")
	}

	// grab totally new copies of the Bytes
	b = s.Bytes()
	b2 = s2.Bytes()

	if bytes.Compare(b2, b) == 0 {
		t.Fatal("segment.Copy.Append mutated original", b, b2)
	}

	n = s.Trunc(256)
	if n != 6 {
		t.Fatal("segment.Trunc(256) didn’t return string length", n)
	}

	n = s.Trunc(4)
	if n != 4 {
		t.Fatal("segment.Trunc(4) didn’t return expected new string length", n)
	}

	// grab totally new copies of the Bytes
	b = s.Bytes()
	b2 = s2.Bytes()

	if bytes.Compare(b, b2) == 0 {
		t.Fatal("mutations to original segment mutated segment.Copy", b, b2)
	}

	if bytes.Compare(b, []byte{0x02, 0x03, 0x05, 0x07}) != 0 {
		t.Fatal("after segment.Trunc, segment.Bytes doesn’t contain expected values", b)
	}

	n = s.Trunc(0)
	if n != 0 {
		t.Fatal("segment.Trunc(0) didn’t return expected new string length", n)
	}

	b = s.Bytes()
	if bytes.Compare(b, []byte{}) != 0 {
		t.Fatal("segment.Trunc(0) doesn’t compare equal to an empty byte slice", b)
	}

	b2 = s2.Bytes()

	if bytes.Compare(b, b2) == 0 {
		t.Fatal("mutations to original segment mutated segment.Copy", b, b2)
	}

	s = s2.Tail(1)
	if s == nil {
		t.Fatal("segment.Tail unexpected returned an empty tail")
	}

	n = s.Len()
	if n != 6 {
		t.Fatal("segment.Tail returned a segment of unexpected length", n)
	}

	b = s.Bytes()
	if bytes.Compare(b, []byte{0x03, 0x05, 0x07, 0x0B, 0x0D, 0x11}) != 0 {
		t.Fatal("after segment.Append, segment contains unexpected content", b)
	}

	n = s.Append(s.Bytes())
	if n != 6 {
		t.Fatal("segment.Append returned unexpected number of bytes appended")
	}

	b = s.Bytes()
	if bytes.Compare(b, []byte{0x03, 0x05, 0x07, 0x0B, 0x0D, 0x11, 0x03, 0x05, 0x07, 0x0B, 0x0D, 0x11}) != 0 {
		t.Fatal("after segment.Append of self, segment contains unexpected content", b)
	}

	// Why not loop this?
	// Each of these results have known expected behavior.
	// By breaking each to its own line, it means we get meaningful line numbers about what went wrong where,
	// rather than relying upon error message reporting to know what went into it being wrong.

	n = s.Append(s.Bytes())
	if n != 12 {
		t.Fatal("segment.Append of self returned unexpected length appended", n)
	}

	n = s.Append(s.Bytes())
	if n != 24 {
		t.Fatal("segment.Append of self returned unexpected length appended", n)
	}

	n = s.Append(s.Bytes())
	if n != 48 {
		t.Fatal("segment.Append of self returned unexpected length appended", n)
	}

	n = s.Append(s.Bytes())
	if n != 96 {
		t.Fatal("segment.Append of self returned unexpected length appended", n)
	}

	n = s.Append(s.Bytes())
	if n != 255-192 {
		// This value is part of the internal contract we are building.
		// If we change the max length/capacity of a segment, we want to break tests,
		// because we will have been breaking assumptions built by contract in the rest of the package.
		t.Fatal("segment.Append of self returned unexpected length appended", n)
	}

	n = s.Append([]byte{0x02})
	if n != 0 {
		t.Fatal("segment.Append to a full buffer returned unexpected length appended", n)
	}
}
