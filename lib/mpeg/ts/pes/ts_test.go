package pes

import (
	"reflect"
	"testing"
)

func TestTimestampEncoding(t *testing.T) {
	var inputTS uint64 = 0
	b := encodeTS(inputTS)
	expected := []byte{0x01, 0x00, 0x01, 0x00, 0x01}

	if !reflect.DeepEqual(b, expected) {
		t.Errorf("encodeTS(0x%09x): expected [% 2x] got [% 2x]", inputTS, expected, b)
	}

	if ts := decodeTS(b); *ts != inputTS {
		t.Errorf("decodeTS·encodeTS: expected 0x%09x got 0x%09x", inputTS, *ts)
	}

	inputTS = 0x1FFFFFFFF
	b = encodeTS(inputTS)
	expected = []byte{0x0f, 0xff, 0xff, 0xff, 0xff}

	if !reflect.DeepEqual(b, expected) {
		t.Errorf("encodeTS(0x%09x): expected [% 2x] got [% 2x]", inputTS, expected, b)
	}

	if ts := decodeTS(b); *ts != inputTS {
		t.Errorf("decodeTS·encodeTS: expected 0x%09x got 0x%09x", inputTS, *ts)
	}

	inputTS = 0x1F3B795D1
	b = encodeTS(inputTS)
	// a bbbbbbbb cccccccc dddddddd eeeeeeee
	// 1 11110011 10110111 10010101 11010001
	// 0000abb1 bbbbbbcc ccccccd1 ddddddde eeeeeee1
	//     111  11001110 1101111  00101011 1010001
	// 00001111 11001110 11011111 00101011 10100011
	//    0   f    c   e    d   f    2   b    a   3
	expected = []byte{0x0f, 0xce, 0xdf, 0x2b, 0xa3}

	if !reflect.DeepEqual(b, expected) {
		t.Errorf("encodeTS(0x%09x): expected [% 2x] got [% 2x]", inputTS, expected, b)
	}

	if ts := decodeTS(b); *ts != inputTS {
		t.Errorf("decodeTS·encodeTS: expected 0x%09x got 0x%09x", inputTS, *ts)
	}
}
