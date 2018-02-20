package pcr

import (
	"fmt"
	"time"
)

type PCR struct {
	base      uint64
	extension uint16
}

func (c *PCR) String() string {
	s := c.Duration().String()

	if c.extension == 0 {
		return s
	}

	return fmt.Sprintf("%s<%d>", s, c.extension)
}

const (
	pcrModulo = (1 << 33) - 1
)

func (c *PCR) Marshal() ([]byte, error) {
	//pcr := uint64(c.base & pcrModulo) << 15 | uint64(c.extension & 0x1ff)

	b := make([]byte, 6)
	b[0] = byte((c.base >> 25) & 0xff)
	b[1] = byte((c.base >> 17) & 0xff)
	b[2] = byte((c.base >> 9) & 0xff)
	b[3] = byte((c.base >> 1) & 0xff)
	b[4] = byte((c.base << 7) & 0x80) | byte((c.extension >> 8) & 0x01)
	b[5] = byte(c.extension & 0xff)

	return b, nil
}

func (c *PCR) Unmarshal(b []byte) error {
	var pcr uint64
	for _, b := range b[0:6] {
		pcr = (pcr << 8) | uint64(b)
	}

	c.base = pcr >> 15
	c.extension = uint16(pcr & 0x1ff)

	return nil
}

func (c *PCR) Duration() time.Duration {
	return (time.Duration(c.base & pcrModulo) * time.Microsecond) / 27
}

func (c *PCR) Set(d time.Duration) {
	c.base = uint64((d * 27) / time.Microsecond) & pcrModulo
}
