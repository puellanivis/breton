package ts

import (
	"bytes"
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

	SpliceCountdown *byte

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
	if af == nil {
		return "{}"
	}

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

	if af.SpliceCountdown != nil {
		out = append(out, fmt.Sprintf("SpliceCountdown:%d", *af.SpliceCountdown))
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

	flagAFExtLTW            = 0x80
	flagAFExtPiecewiseRate  = 0x40
	flagAFExtSeamlessSplice = 0x20

	flagAFExtLTWValid = 0x80

	adaptationFieldMinLength = 2
)

func (af *AdaptationField) len() int {
	if af == nil {
		return 0
	}

	l := 2

	if af.PCR != nil {
		l += 6
	}

	if af.OPCR != nil {
		l += 6
	}

	if af.SpliceCountdown != nil {
		l++
	}

	l += len(af.PrivateData)

	if af.LegalTimeWindow.Value != nil || af.PiecewiseRate != nil || af.SeamlessSplice.DTS != nil {
		l += 2

		if af.LegalTimeWindow.Value != nil {
			l += 2
		}

		if af.PiecewiseRate != nil {
			l += 3
		}

		if af.SeamlessSplice.DTS != nil {
			l += 5
		}
	}

	return l + af.Stuffing
}

func (af *AdaptationField) marshal() ([]byte, error) {
	if af == nil {
		// If we got here, we already set that there is an AdaptationField…
		// If so, return an empty AdaptationField, not a not-there AdaptationField.
		return []byte{1, 0}, nil
	}

	b := make([]byte, 2)

	if af.Discontinuity {
		b[1] |= flagAFDiscontinuity
	}
	if af.RandomAccess {
		b[1] |= flagAFRandomAccess
	}
	if af.Priority {
		b[1] |= flagAFPriority
	}

	if af.PCR != nil {
		b[1] |= flagAFPCR

		pcr, err := af.PCR.Marshal()
		if err != nil {
			return nil, err
		}

		b = append(b, pcr...)
	}

	if af.OPCR != nil {
		b[1] |= flagAFOPCR

		opcr, err := af.OPCR.Marshal()
		if err != nil {
			return nil, err
		}

		b = append(b, opcr...)
	}

	if af.SpliceCountdown != nil {
		b[1] |= flagAFSplicePoint
		b = append(b, *af.SpliceCountdown)
	}

	if af.PrivateData != nil {
		if len(af.PrivateData) > 0xFF {
			return nil, errors.Errorf("private_data length exceeds 255: %d", len(af.PrivateData))
		}

		b[1] |= flagAFPrivateData
		b = append(b, byte(len(af.PrivateData)&0xFF))
		b = append(b, af.PrivateData...)
	}

	if af.LegalTimeWindow.Value != nil || af.PiecewiseRate != nil || af.SeamlessSplice.DTS != nil {
		ext := make([]byte, 2)

		if af.LegalTimeWindow.Value != nil {
			ext[1] |= flagAFExtLTW

			ext = append(ext,
				byte((*af.LegalTimeWindow.Value>>8)&0x7F),
				byte(*af.LegalTimeWindow.Value&0xFF),
			)

			if af.LegalTimeWindow.Valid {
				ext[2] |= flagAFExtLTWValid
			}
		}

		if af.PiecewiseRate != nil {
			ext[1] |= flagAFExtPiecewiseRate

			ext = append(ext,
				byte((*af.PiecewiseRate>>16)&0x3F),
				byte((*af.PiecewiseRate>>8)&0xFF),
				byte(*af.PiecewiseRate&0xFF),
			)
		}

		if af.SeamlessSplice.DTS != nil {
			ext[1] |= flagAFExtSeamlessSplice

			ext = append(ext,
				byte((af.SeamlessSplice.Type&0x0F)<<4)|byte((*af.SeamlessSplice.DTS>>29)&0xE)|1,
				byte((*af.SeamlessSplice.DTS>>23)&0xFF),
				byte((*af.SeamlessSplice.DTS>>14)&0xFE)|1,
				byte((*af.SeamlessSplice.DTS>>7)&0xFF),
				byte((*af.SeamlessSplice.DTS<<1)&0xFE)|1,
			)
		}

		ext[0] = byte(len(ext) - 1)

		b = append(b, ext...)
	}

	if af.Stuffing > 0 {
		b = append(b, bytes.Repeat([]byte{0xFF}, af.Stuffing)...)
	}

	b[0] = byte(len(b) - 1)
	return b, nil
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
		sc := b[start]
		af.SpliceCountdown = &sc
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

		if ext[0]&flagAFExtLTW != 0 {
			b := ext[start:]

			af.LegalTimeWindow.Valid = b[0]&0x80 != 0

			ltw := uint16(b[0]&0x7F)<<8 | uint16(b[1])
			af.LegalTimeWindow.Value = &ltw

			start += 2
		}

		if ext[0]&flagAFExtPiecewiseRate != 0 {
			b := ext[start:]

			pwr := uint32(b[0]&0x3F)<<16 | uint32(b[1])<<8 | uint32(b[2])
			af.PiecewiseRate = &pwr

			start += 3
		}

		if ext[0]&flagAFExtSeamlessSplice != 0 {
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
