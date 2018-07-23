package psi

import (
	"github.com/pkg/errors"
)

// CommonMarshal takes care of the common elements of encoding a PSI table.
func CommonMarshal(id uint8, private bool, syn *SectionSyntax, data []byte) ([]byte, error) {
	// len(common_header) + len(data) + len(crc)
	l := 4 + len(data) + 4

	var sb []byte
	if syn != nil {
		var err error

		sb, err = syn.Marshal()
		if err != nil {
			return nil, err
		}

		l += len(sb)
	}

	b := make([]byte, l)

	b[1] = id
	if private {
		b[2] |= flagPrivate
	}

	start := 4
	secLen := l - start
	if secLen > 1021 {
		return nil, errors.New("section_length may not exceed 1021")
	}
	b[2] |= byte((secLen>>8)&0x0F) | 0x30
	b[3] = byte(secLen & 0xFF)

	if syn != nil {
		b[2] |= flagSectionSyntax

		copy(b[start:], sb)
		start += len(sb)
	}

	// copy in the SectionSyntax
	copy(b[start:], data)

	// TODO(puellanivis): calculate CRC32
	copy(b[l-4:], []byte{0xde, 0xad, 0xbe, 0xef})

	return b, nil
}

// CommonUnmarshal takes care of the common elements of decoding a PSI table.
func CommonUnmarshal(b []byte) (syn *SectionSyntax, data []byte, crc uint32, err error) {
	secLen := int(b[1]&0x0F)<<8 | int(b[2])
	if secLen > 1021 {
		return nil, nil, 0, errors.New("section_length may not exceed 1021")
	}

	start := 3

	if b[1]&flagSectionSyntax != 0 {
		syn = new(SectionSyntax)

		if err := syn.Unmarshal(b[3:]); err != nil {
			return nil, nil, 0, err
		}

		start += 5
		secLen -= 5
	}

	end := start + secLen - 4
	if start >= len(b) {
		return nil, nil, 0, errors.New("buffer too short")
	}
	if end+4 > len(b) {
		return nil, nil, 0, errors.New("section_length overruns buffer")
	}
	if end < 0 {
		return nil, nil, 0, errors.New("section_length is too short")
	}
	data = b[start:end]

	for _, b := range b[end : end+4] {
		crc = (crc << 8) | uint32(b)
	}

	return syn, data, crc, nil
}
