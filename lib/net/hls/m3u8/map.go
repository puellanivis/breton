package m3u8

// Map implements the MAP directive of the m3u8 standard.
type Map struct {
	URI       string    `m3u8:"URI"`
	ByteRange ByteRange `m3u8:"BYTERANGE"`
}
