package descriptor

import ()

var (
	descriptorRegistry = make(map[uint8]func() Descriptor)
)

func Register(tag uint8, fn func() Descriptor) {
	descriptorRegistry[tag] = fn
}

type Descriptor interface {
	Tag() uint8
	Len() int

	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

func defaultDescriptor() Descriptor {
	return new(raw)
}

func Unmarshal(b []byte) (d Descriptor, err error) {
	fn := descriptorRegistry[uint8(b[0])]
	if fn == nil {
		fn = defaultDescriptor
	}

	d = fn()

	return d, d.Unmarshal(b)
}
