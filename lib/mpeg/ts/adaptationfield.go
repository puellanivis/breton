package ts

import (
	"fmt"
	"strings"
	
	"github.com/pkg/errors"
	"github.com/puellanivis/breton/lib/mpeg/ts/pcr"
)

type AdaptationField struct {
	Discontinuity bool
	RandomAccess  bool
	Priority      bool

	PCR  *pcr.PCR
	OPCR *pcr.PCR

	SpliceCountdown byte

	PrivateData []byte

	LegalTimeWindow struct {
		Valid bool
		Value *uint16
	}

	PiecewiseRate *uint32

	SeamlessSplice struct {
		Type uint8
		DTS  *uint64
	}

	Stuffing int
}

func (af *AdaptationField) String() string {
	var out []string

	if af.Discontinuity {
		out = append(out, "DISCONT")
	}

	if af.RandomAccess {
		out = append(out, "RAND")
	}

	if af.Priority {
		out = append(out, "PRI")
	}

	if af.PCR != nil {
		out = append(out, fmt.Sprintf("PCR:%v", af.PCR))
	}

	if af.OPCR != nil {
		out = append(out, fmt.Sprintf("OPCR:%v", af.OPCR))
	}

	if af.SpliceCountdown > 0 {
		out = append(out, fmt.Sprintf("SpliceCountdown:%d", af.SpliceCountdown))
	}

	if len(af.PrivateData) > 0 {
		out = append(out, fmt.Sprintf("priv[%d]", len(af.PrivateData)))
	}

	if af.LegalTimeWindow.Value != nil {
		valid := ""
		if af.LegalTimeWindow.Valid {
			valid = "(VALID)"
		}
		out = append(out, fmt.Sprintf("LTW%s:x%04X", valid, *af.LegalTimeWindow.Value))
	}

	if af.PiecewiseRate != nil {
		out = append(out, fmt.Sprintf("PWR:x%06X", *af.PiecewiseRate))
	}

	if af.SeamlessSplice.DTS != nil {
		out = append(out, fmt.Sprintf("SeamlessSplice(%02X):x%09X", af.SeamlessSplice.Type, *af.SeamlessSplice.DTS))
	}

	if af.Stuffing > 0 {
		out = append(out, fmt.Sprintf("Stuffing[%d]", af.Stuffing))
	}

	return fmt.Sprintf("{%s}", strings.Join(out, " "))
}

const (
	flagAFDiscontinuity = 0x80
	flagAFRandomAccess  = 0x40
	flagAFPriority      = 0x20
	flagAFPCR           = 0x10
	flagAFOPCR          = 0x08
	flagAFSplicePoint   = 0x04
	flagAFPrivateData   = 0x02
	flagAFExtension     = 0x01

	flagAFELTW            = 0x80
	flagAFEPiecewiseRate  = 0x40
	flagAFESeamlessSplice = 0x20
)

func (af *AdaptationField) marshal() ([]byte, error) {
	return nil, errors.New("unimplemented")
}

func (af *AdaptationField) unmarshal(b []byte) (int, error) {
	if b[0] == 0 {
		return 1, nil
	}

	length := int(b[0]) + 1

	// trim so that OOB access will panic
	b = b[:length]

	af.Discontinuity = b[1]&flagAFDiscontinuity != 0
	af.RandomAccess = b[1]&flagAFRandomAccess != 0
	af.Priority = b[1]&flagAFPriority != 0

	start := 2

	if b[1]&flagAFPCR != 0 {
		af.PCR = new(pcr.PCR)

		_ = af.PCR.Unmarshal(b[start : start+6])

		start += 6
	}

	if b[1]&flagAFOPCR != 0 {
		af.OPCR = new(pcr.PCR)

		_ = af.OPCR.Unmarshal(b[start : start+6])

		start += 6
	}

	if b[1]&flagAFSplicePoint != 0 {
		af.SpliceCountdown = b[start]
		start++
	}

	if b[1]&flagAFPrivateData != 0 {
		l := int(b[start])
		start++

		af.PrivateData = append([]byte{}, b[start:start+l]...)
		start += l
	}

	if b[1]&flagAFExtension != 0 {
		l := int(b[start])
		start++

		ext := b[start : start+l]
		start += l

		if ext[0]&flagAFELTW != 0 {
			b := ext[start:]

			af.LegalTimeWindow.Valid = b[0]&0x80 != 0

			ltw := uint16(b[0]&0x7F)<<8 | uint16(b[1])
			af.LegalTimeWindow.Value = &ltw

			start += 2
		}

		if ext[0]&flagAFEPiecewiseRate != 0 {
			b := ext[start:]

			pwr := uint32(b[0]&0x3F)<<16 | uint32(b[1])<<8 | uint32(b[2])
			af.PiecewiseRate = &pwr

			start += 3
		}

		if ext[0]&flagAFESeamlessSplice != 0 {
			b := ext[start:]

			af.SeamlessSplice.Type = (b[0] >> 4) & 0x0F

			ts := uint64(b[start]>>1) & 0x07
			ts = (ts << 8) | uint64(b[1])
			ts = (ts << 7) | uint64((b[2]>>1)&0x7F)
			ts = (ts << 8) | uint64(b[3])
			ts = (ts << 7) | uint64((b[4]>>1)&0x7F)

			af.SeamlessSplice.DTS = &ts

			start += 5
		}
	}

	af.Stuffing = len(b) - start

	return length, nil
}
