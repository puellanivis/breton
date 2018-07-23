package descriptor

import (
	"github.com/pkg/errors"
)

var (
	descriptorRegistry = make(map[uint8]func() Descriptor)
)

// Register registers a new tag to a function that will return an appropriate Descriptor for that tag.
func Register(tag uint8, fn func() Descriptor) {
	descriptorRegistry[tag] = fn
}

// Descriptor defines an MPEG-TS Descriptor.
type Descriptor interface {
	Tag() uint8
	Len() int

	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

func defaultDescriptor() Descriptor {
	return new(raw)
}

// Unmarshal decodes an arbitrary Descriptor from a byte slice.
func Unmarshal(b []byte) (d Descriptor, err error) {
	if len(b) < 1 {
		return nil, errors.New("empty buffer")
	}

	fn := descriptorRegistry[uint8(b[0])]
	if fn == nil {
		fn = defaultDescriptor
	}

	d = fn()

	return d, d.Unmarshal(b)
}
