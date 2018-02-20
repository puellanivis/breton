package descriptor

import ()

type Descriptor interface {
	Tag() uint8
	Len() int

	// Marshal and Unmarshal produce the payload
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

func Unmarshal(b []byte) (d Descriptor, err error) {
	switch b[0] {
	case tagDVBService:
		d = new(DVBService)

	default:
		d = new(raw)
	}

	return d, d.Unmarshal(b)
}
