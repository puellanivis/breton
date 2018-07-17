package m3u8

type Define struct {
	Name  string `m3u8:"NAME,optional"`
	Value string `m3u8:"VALUE,optional"`

	Import string `m3u8:"IMPORT,optional"`
}
