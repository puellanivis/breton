package m3u8

import (
	"fmt"
	"strconv"
	"strings"
)

type ByteRange struct {
	Length int
	Offset int
}

func (r *ByteRange) UnmarshalString(value string) error {
	fields := strings.Split(value, "@")

	var offset int

	switch len(fields) {
	case 1:
	case 2:
		var err error
		offset, err = strconv.Atoi(fields[1])
		if err != nil {
			return fmt.Errorf("invalid offset in BYTERANGE: %q: %v", fields[1], err)
		}

	default:
		return fmt.Errorf("invalid BYTERANGE: %q", value)
	}

	length, err := strconv.Atoi(fields[0])
	if err != nil {
		return fmt.Errorf("invalid length in BYTERANGE: %q: %v", fields[0], err)
	}

	r.Length = length
	r.Offset = offset

	return nil
}

func (r ByteRange) String() string {
	if r.Offset != 0 {
		return fmt.Sprintf("%d@%d", r.Length, r.Offset)
	}

	return strconv.Itoa(r.Length)
}
