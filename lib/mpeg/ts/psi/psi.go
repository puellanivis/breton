package psi

import ()

var tableRegistry = make(map[uint8]func() PSI)

func Register(id uint8, fn func() PSI) {
	tableRegistry[id] = fn
}

type PSI interface {
	TableID() uint8
	SectionSyntax() *SectionSyntax

	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

func defaultTable() PSI {
	return new(raw)
}

func Unmarshal(b []byte) (psi PSI, err error) {
	ptrVal := int(b[0])
	b = b[1+ptrVal:]

	fn := tableRegistry[uint8(b[0])]
	if fn == nil {
		fn = defaultTable
	}

	psi = fn()

	return psi, psi.Unmarshal(b)
}
