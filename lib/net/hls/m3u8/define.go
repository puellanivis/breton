package m3u8

// Define implements the DEFINE m3u8 directive.
type Define struct {
	Name  string `m3u8:"NAME,optional"`
	Value string `m3u8:"VALUE,optional"`

	Import string `m3u8:"IMPORT,optional"`
}
