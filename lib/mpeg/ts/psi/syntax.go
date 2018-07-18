package psi

import (
	"fmt"
	"strings"
)

// SectionSyntax defines the MPEG-TS Section Syntax which is common to all PSI tables.
type SectionSyntax struct {
	TableIDExtension  uint16
	Version           uint8
	Current           bool
	SectionNumber     uint8
	LastSectionNumber uint8
}

const (
	shiftSyntaxVersion = 1
	maskSyntaxVersion  = 0x1F

	flagSyntaxCurrent = 0x01
)

func (s *SectionSyntax) String() string {
	out := []string{
		fmt.Sprintf("TIE:x%04X", s.TableIDExtension),
		fmt.Sprintf("VER:%X", s.Version),
	}

	if s.Current {
		out = append(out, "CUR")
	}

	if s.SectionNumber|s.LastSectionNumber != 0 {
		out = append(out, fmt.Sprintf("SecNum:x%02X LastSec:x%02X", s.SectionNumber, s.LastSectionNumber))
	}

	return fmt.Sprintf("{%s}", strings.Join(out, " "))
}

// Unmarshal decodes a byte slice into the SectionSyntax.
func (s *SectionSyntax) Unmarshal(b []byte) error {
	s.TableIDExtension = (uint16(b[0]) << 8) | uint16(b[1])
	s.Version = (b[2] >> shiftSyntaxVersion) & maskSyntaxVersion
	s.Current = b[2]&flagSyntaxCurrent != 0
	s.SectionNumber = b[3]
	s.LastSectionNumber = b[4]

	return nil
}

// Marshal encodes the SectionSyntax into a byte slice.
func (s *SectionSyntax) Marshal() ([]byte, error) {
	b := make([]byte, 5)

	b[0] = byte((s.TableIDExtension >> 8) & 0xFF)
	b[1] = byte(s.TableIDExtension & 0xFF)
	b[2] = 0xC0 | (s.Version&maskSyntaxVersion)<<shiftSyntaxVersion

	if s.Current {
		b[2] |= flagSyntaxCurrent
	}

	b[3] = s.SectionNumber
	b[4] = s.LastSectionNumber

	return b, nil
}
