package m3u8

import (
	"time"
)

// MasterPlaylist implements the primary playlist of the m3u8 standard.
type MasterPlaylist struct {
	Version int `m3u8:"EXT-X-VERSION"`

	Media *Media `m3u8:"EXT-X-MEDIA,attribute-list"`

	StreamInf       *StreamInf       `m3u8:"EXT-X-STREAM-INF,attribute-list"`
	IFrameStreamInf *IFrameStreamInf `m3u8:"EXT-X-I-FRAME-STREAM-INF,attribute-list"`

	SessionData *SessionData `m3u8:"EXT-X-SESSION-DATA,attribute-list"`
	SessionKey  *Key         `m3u8:"EXT-X-SESSION-KEY,attribute-list"`

	IndependentSegments bool `m3u8:"EXT-X-INDEPENDENT-SEGMENTS,optional"`

	Start  *Start   `m3u8:"EXT-X-START,attribute-list"`
	Define []Define `m3u8:"EXT-X-DEFINE,attribute-list"`
}

// MediaPlaylist implements the media-specific playlist of the m3u8 standard.
type MediaPlaylist struct {
	Version int `m3u8:"EXT-X-VERSION"`

	TargetDuration time.Duration `m3u8:"EXT-X-TARGETDURATION"`

	MediaSequence         int `m3u8:"EXT-X-MEDIA-SEQUENCE,optional"`
	DiscontinuitySequence int `m3u8:"EXT-X-DISCONTINUITY-SEQUENCE,optional"`

	EndList     bool `m3u8:"EXT-X-ENDLIST,optional"`
	IFramesOnly bool `m3u8:"EXT-X-I-FRAMES-ONLY,version=4,optional"`

	PlaylistType string `m3u8:"EXT-X-PLAYLIST-TYPE,optional" enum:"EVENT,VOD"`

	IndependentSegments bool `m3u8:"EXT-X-INDEPENDENT-SEGMENTS,optional"`

	Start  *Start   `m3u8:"EXT-X-START,attribute-list"`
	Define []Define `m3u8:"EXT-X-DEFINE,attribute-list"`
}

// MediaSegment implements the MEDIA-SEGMENT directive of the m3u8 standard.
type MediaSegment struct {
	// Encoded by EXTINF, tell the Marshaller to ignore them.
	Duration float64 `m3u8:"-"`
	Title    string  `m3u8:"-"`

	ByteRange ByteRange `m3u8:"EXT-X-BYTERANGE,version=4,optional"`

	Discontinuity bool `m3u8:"EXT-X-DISCONTINUITY,optional"`
	Gap           bool `m3u8:"EXT-X-GAP,optional"`

	Key *Key `m3u8:"EXT-X-KEY,attribute-list"`
	Map *Map `m3u8:"EXT-X-MAP,attribute-list"`

	ProgramDateTime time.Time `m3u8:"EXT-X-PROGRAM-DATE-TIME,optional" format:"2006-01-02T15:04:05.999Z07:00"`

	DateRange *DateRange `m3u8:"EXT-X-DATERANGE,attribute-list"`

	URL string `m3u8:""`
}
