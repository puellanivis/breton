package psi

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type ProgramMap struct {
	ProgramNumber uint16
	PID           uint16
}

type PAT struct {
	Syntax *SectionSyntax

	Map []ProgramMap

	crc uint32
}

func (pat *PAT) String() string {
	out := []string{
		"PAT",
	}

	if pat.Syntax != nil {
		out = append(out, fmt.Sprint(pat.Syntax))
	}

	for _, m := range pat.Map {
		out = append(out, fmt.Sprintf("x%X:x%X", m.ProgramNumber, m.PID))
	}

	out = append(out, fmt.Sprintf("crc:x%08X", pat.crc))

	return fmt.Sprintf("{%s}", strings.Join(out, " "))
}

const (
	tableidPAT = 0x00
)

func init() {
	Register(tableidPAT, func() PSI { return new(PAT) })
}

func (pat *PAT) TableID() uint8 {
	return tableidPAT
}

func (pat *PAT) SectionSyntax() *SectionSyntax {
	return pat.Syntax
}

func (pat *PAT) Unmarshal(b []byte) error {
	if b[0] != tableidPAT {
		return errors.Errorf("table_id mismatch: x%02X != x%02X", b[0], tableidPAT)
	}

	syn, data, crc, err := CommonUnmarshal(b)
	if err != nil {
		return err
	}

	pat.Syntax = syn
	pat.crc = crc

	pat.Map = make([]ProgramMap, len(data)/4)
	for i := range pat.Map {
		b := data[i*4:]

		pat.Map[i].ProgramNumber = (uint16(b[0]) << 8) | uint16(b[1])
		pat.Map[i].PID = (uint16(b[2]&0x1F) << 8) | uint16(b[3])
	}

	return nil
}

func (pat *PAT) Marshal() ([]byte, error) {
	data := make([]byte, len(pat.Map)*4)

	for i := range pat.Map {
		b := data[i*4:]

		b[0] = byte((pat.Map[i].ProgramNumber >> 8) & 0xFF)
		b[1] = byte(pat.Map[i].ProgramNumber & 0xFF)

		b[2] = byte((pat.Map[i].PID >> 8) & 0x1F)
		b[3] = byte(pat.Map[i].PID & 0xFF)
	}

	return CommonMarshal(tableidPAT, false, pat.Syntax, data)
}
