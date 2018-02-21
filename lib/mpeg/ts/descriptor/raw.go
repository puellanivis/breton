package descriptor

import (
	"fmt"

	"github.com/pkg/errors"
)

type raw struct {
	tag  uint8
	data []byte
}

func (d *raw) String() string {
	return fmt.Sprintf("{tag:x%02X data[%d]}", d.tag, len(d.data))
}

func (d *raw) Tag() uint8 {
	return d.tag
}

func (d *raw) Len() int {
	return 2 + len(d.data)
}

func (d *raw) Unmarshal(b []byte) error {
	d.tag = uint8(b[0])
	l := int(b[1])

	b = b[2:]

	if len(b) < l {
		return errors.Errorf("unexpected end of byte-slice: %d < %d", len(b), l)
	}

	d.data = make([]byte, l)
	copy(d.data, b)

	return nil
}

func (d *raw) Marshal() ([]byte, error) {
	if len(d.data) > 0xff {
		return nil, errors.Errorf("descriptor data field too large: %d", len(d.data))
	}

	b := make([]byte, d.Len())

	b[0] = byte(d.tag)
	b[1] = byte(len(d.data))
	copy(b[2:], d.data)

	return b, nil
}
