package psi

var tableRegistry = make(map[uint8]func() PSI)

// Register sets a mapping from the id to a function which returns a PSI of the appropriate type for that id.
func Register(id uint8, fn func() PSI) {
	tableRegistry[id] = fn
}

// PSI is a Program-Specific-Information table.
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
