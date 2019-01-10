package pes

// Timestamp returns a pointer to the given argument.
func Timestamp(ts uint64) *uint64 {
	return &ts
}

func decodeTS(b []byte) *uint64 {
	ts := uint64(b[0]>>1) & 0x07
	ts = (ts << 8) | uint64(b[1])
	ts = (ts << 7) | uint64((b[2]>>1)&0x7F)
	ts = (ts << 8) | uint64(b[3])
	ts = (ts << 7) | uint64((b[4]>>1)&0x7F)

	return &ts
}

func encodeTS(ts uint64) []byte {
	b := make([]byte, 5)

	b[0] = byte((ts>>29)&0x0E) | 1
	b[1] = byte((ts >> 22) & 0xFF)
	b[2] = byte((ts>>14)&0xFE) | 1
	b[3] = byte((ts >> 7) & 0xFF)
	b[4] = byte((ts<<1)&0xFE) | 1

	return b
}
