package m3u8

type Map struct {
	URI       string    `m3u8:"URI"`
	ByteRange ByteRange `m3u8:"BYTERANGE"`
}
