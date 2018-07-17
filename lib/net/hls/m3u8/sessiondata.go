package m3u8

type SessionData struct {
	DataID   string `m3u8:"DATA-ID"`
	Value    string `m3u8:"VALUE,optional"`
	URI      string `m3u8:"URI,optional"`
	Language string `m3u8:"LANGUAGE,optional"`
}
