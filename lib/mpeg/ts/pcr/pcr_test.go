package pcr

import (
	"reflect"
	"testing"
	"time"
)

func TestMarshaling(t *testing.T) {
	var pcr PCR

	pcr.Set(1 * time.Microsecond)

	if pcr.base != 27 {
		t.Errorf("expected pcr.base == 27 but got %d", pcr.base)
	}

	// a bbbbbbbb cccccccc dddddddd eeeeeeee
	// 0 00000000 00000000 00000000 00011011
	//   abbbbbbb bccccccc cddddddd deeeeeee e0000000 00000000
	//   00000000 00000000 00000000 00001101 1
	//   00000000 00000000 00000000 00001101 10000000 00000000
	//      0   0    0   0    0   0    0   d    8   0    0   0

	expected := []byte{ 0x00, 0x00, 0x00, 0x0d, 0x80, 0x00 }
	b, err := pcr.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(b, expected) {
		t.Errorf("pcr.Marshal: expected [% 2x], got [% 2x]", expected, b)
	}

	pcr.base = pcrModulo
	pcr.extension = 0x1ff

	// a bbbbbbbb cccccccc dddddddd eeeeeeee        f gggggggg
	// 1 11111111 11111111 11111111 11111111        1 11111111
	//   abbbbbbb bccccccc cddddddd deeeeeee e000000f gggggggg
	//   11111111 11111111 11111111 11111111 1      1 11111111
	//   11111111 11111111 11111111 11111111 10000001 11111111
	//      f   f    f   f    f   f    f   f    8   1    f   f

	expected = []byte{ 0xff, 0xff, 0xff, 0xff, 0x81, 0xff }
	b, err = pcr.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(b, expected) {
		t.Errorf("pcr.Marshal: expected [% 2x], got [% 2x]", expected, b)
	}

	pcr.base = 0x1F3B795D1
	pcr.extension = 0x15B

	// a bbbbbbbb cccccccc dddddddd eeeeeeee        f gggggggg
	// 1 11110011 10110111 10010101 11010001        1 01011011
	//   abbbbbbb bccccccc cddddddd deeeeeee e000000f gggggggg
	//   11111001 11011011 11001010 11101000 1      1 01011011
	//   11111001 11011011 11001010 11101000 10000001 01011011
	//      f   9    d   b    c   a    e   8    8   1    5   b

	expected = []byte{ 0xf9, 0xdb, 0xca, 0xe8, 0x81, 0x5b }
	b, err = pcr.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(b, expected) {
		t.Errorf("pcr.Marshal: expected [% 2x], got [% 2x]", expected, b)
	}

}
