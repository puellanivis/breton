package m3u8

import ()

type Media struct {
	Type    string `m3u8:"TYPE" enum:"AUDIO,VIDEO,SUBTITLES,CLOSED-CAPTIONS"`
	GroupID string `m3u8:"GROUP-ID"`
	Name    string `m3u8:"NAME"`

	Default    bool `m3u8:"DEFAULT,optional" enum:"NO,YES"`
	Autoselect bool `m3u8:"AUTOSELECT,optional" enum:"NO,YES"`
	Forced     bool `m3u8:"FORCED,optional" enum:"NO,YES"`

	Language      string `m3u8:"LANGUAGE,optional"`
	AssocLanguage string `m3u8:"ASSOC-LANGUAGE,optional"`

	InstreamID      string `m3u8:"INSTREAM-ID,optional"`
	Characteristics string `m3u8:"CHARACTERISTICS,optional"`

	// Standard does not technically require that all parameters be decimal-integers.
	Channels []int `m3u8:"CHANNELS,optional" delim:"/"`

	URI string `m3u8:"URI,optional"`
}
