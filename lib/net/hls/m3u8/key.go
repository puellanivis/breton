package m3u8

type Key struct {
	Method string `m3u8:"METHOD" enum:"NONE,AES-128,SAMPLE-AES"`
	URI    string `m3u8:"URI,optional"`

	InitializationVector []byte `m3u8:"IV,optional,version=2"`
	KeyFormatVersions    []int  `m3u8:"KEYFORMATVERSIONS,optional,version=5" delim:"/"`
}
