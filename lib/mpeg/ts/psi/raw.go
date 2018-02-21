package psi

import (
	"fmt"
	"strings"
)

type raw struct {
	ID      byte
	Private bool

	Syntax *SectionSyntax

	Data []byte

	crc uint32
}

func (psi *raw) TableID() uint8 {
	return psi.ID
}

func (psi *raw) SectionSyntax() *SectionSyntax {
	return psi.Syntax
}

func (psi *raw) String() string {
	var out []string

	out = append(out, fmt.Sprintf("TID:x%02X", psi.ID))
	if psi.Private {
		out = append(out, "PRIV")
	}

	if psi.Syntax != nil {
		out = append(out, fmt.Sprint(psi.Syntax))
	}

	out = append(out, fmt.Sprintf("Data[%d]", len(psi.Data)))

	out = append(out, fmt.Sprintf("crc:x%08X", psi.crc))

	return fmt.Sprintf("{%s}", strings.Join(out, " "))
}

const (
	flagSectionSyntax = 0x80
	flagPrivate       = 0x40
)

func (psi *raw) Unmarshal(b []byte) error {
	psi.ID = b[0]
	psi.Private = b[1]&flagPrivate != 0

	syn, data, crc, err := CommonUnmarshal(b)
	if err != nil {
		return err
	}

	psi.Syntax = syn
	psi.crc = crc

	psi.Data = append([]byte{}, data...)

	return nil
}

func (psi *raw) Marshal() ([]byte, error) {
	return CommonMarshal(psi.ID, psi.Private, psi.Syntax, psi.Data)
}
