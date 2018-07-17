package m3u8

import (
	"bytes"
	"fmt"
	"strconv"
)

// ByteRange implements the BYTERANGE attribute of the m3u8 format.
type ByteRange struct {
	Length int
	Offset int
}

// TextUnmarshal implements encoding.TextUnmarshaler
func (r *ByteRange) TextUnmarshal(value []byte) error {
	fields := bytes.Split(value, []byte{'@'})

	var offset int

	switch len(fields) {
	case 1:
	case 2:
		var err error
		offset, err = strconv.Atoi(string(fields[1]))
		if err != nil {
			return fmt.Errorf("invalid offset in BYTERANGE: %q: %v", fields[1], err)
		}

	default:
		return fmt.Errorf("invalid BYTERANGE: %q", value)
	}

	length, err := strconv.Atoi(string(fields[0]))
	if err != nil {
		return fmt.Errorf("invalid length in BYTERANGE: %q: %v", fields[0], err)
	}

	r.Length = length
	r.Offset = offset

	return nil
}

// TextMarshal implements encoding.TextMarshal.
func (r ByteRange) TextMarshal() ([]byte, error) {
	return []byte(r.String()), nil
}

// String implements fmt.Stringer.
func (r ByteRange) String() string {
	if r.Offset != 0 {
		return fmt.Sprintf("%d@%d", r.Length, r.Offset)
	}

	return strconv.Itoa(r.Length)
}
