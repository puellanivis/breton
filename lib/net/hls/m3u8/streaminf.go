package m3u8

import ()

type StreamInf struct {
	Bandwidth        int        `m3u8:"BANDWIDTH"`
	AverageBandwidth int        `m3u8:"AVERAGE-BANDWIDTH,optional"`
	Codecs           []string   `m3u8:"CODECS,optional" delim:","`
	Resolution       Resolution `m3u8:"RESOLUTION,optional"`
	HDCPLevel        string     `m3u8:"HDCP-LEVEL,optional" enum:"NONE,TYPE-0,TYPE-1"`
	VideoRange       string     `m3u8:"VIDEO-RANGE,optional" enum:"SDR,PQ"`

	FrameRate float64 `m3u8:"FRAME-RATE,optional"`

	Audio          string `m3u8:"AUDIO,optional"`
	Subtitles      string `m3u8:"SUBTITLES,optional"`
	ClosedCaptions string `m3u8:"CLOSED-CAPTIONS,optional"`
}

type IFrameStreamInf struct {
	Bandwidth        int        `m3u8:"BANDWIDTH"`
	AverageBandwidth int        `m3u8:"AVERAGE-BANDWIDTH,optional"`
	Codecs           []string   `m3u8:"CODECS,optional" delim:","`
	Resolution       Resolution `m3u8:"RESOLUTION,optional"`
	HDCPLevel        string     `m3u8:"HDCP-LEVEL,optional" enum:"NONE,TYPE-0,TYPE-1"`
	VideoRange       string     `m3u8:"VIDEO-RANGE,optional" enum:"SDR,PQ"`

	URI string `m3u8:"URI"`
}
