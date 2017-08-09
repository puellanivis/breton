package mpd

import (
	//"encoding/base64"
	//"encoding/hex"
	//"encoding/xml"
	//"errors"
	"regexp"
	//"strings"
	"time"
)

// mostly computer generated from from http://standards.iso.org/ittf/PubliclyAvailableStandards/MPEG-DASH_schema_files/DASH-MPD.xsd

type Presentation string

var Presentation_Valid = map[string]bool{
	"static":  true,
	"dynamic": true,
}

type Ratio string

var Ratio_Validate = regexp.MustCompile(`[0-9]*:[0-9]*`)

type FrameRate string

var FrameRate_Validate = regexp.MustCompile(`[0-9]*[0-9](/[0-9]*[0-9])?`)

type ConditionalUint string // union {
// uint
// bool
// }

type StringNoWhitespace string

var StringNoWhitespace_Validate = regexp.MustCompile(`[^\r\n\t \p{Z}]*`)

type SAP uint

const (
	SAP_MinInclusive SAP = 0
	SAP_MaxInclusive SAP = 6
)

type VideoScan string

var VideoScan_Valid = map[string]bool{
	"progressive": true,
	"interlaced":  true,
	"unknown":     true,
}

type StringVector []string

type UIntVector []uint

type MPD struct {
	ProgramInformation   []*ProgramInformation `xml:"ProgramInformation,omitempty"`
	BaseURL              []*BaseURL            `xml:"BaseURL,omitempty"`
	Location             []string              `xml:"Location,omitempty"`
	Period               []*Period             `xml:"Period"`
	Metrics              []*Metrics            `xml:"Metrics,omitempty"`
	EssentialProperty    []*Descriptor         `xml:"EssentialProperty,omitempty"`
	SupplementalProperty []*Descriptor         `xml:"SupplementalProperty,omitempty"`
	UTCTiming            []*Descriptor         `xml:"UTCTiming,omitempty"`

	Id                         string       `xml:"id,attr,omitempty"`
	Profiles                   string       `xml:"profiles,attr"`
	Type                       Presentation `xml:"type,attr,omitempty"` // default: static
	AvailabilityStartTime      time.Time    `xml:"availabilityStartTime,attr,omitempty"`
	AvailabilityEndTime        time.Time    `xml:"availabilityEndTime,attr,omitempty"`
	PublishTime                time.Time    `xml:"publishTime,attr,omitempty"`
	MediaPresentationDuration  Duration     `xml:"mediaPresentationDuration,attr,omitempty"`
	MinimumUpdatePeriod        Duration     `xml:"minimumUpdatePeriod,attr,omitempty"`
	MinBufferTime              Duration     `xml:"minBufferTime,attr"`
	TimeShiftBufferDepth       Duration     `xml:"timeShiftBufferDepth,attr,omitempty"`
	SuggestedPresentationDelay Duration     `xml:"suggestedPresentationDelay,attr,omitempty"`
	MaxSegmentDuration         Duration     `xml:"maxSegmentDuration,attr,omitempty"`
	MaxSubsegmentDuration      Duration     `xml:"maxSubsegmentDuration,attr,omitempty"`
}

type Period struct {
	BaseURL              []*BaseURL       `xml:"BaseURL,omitempty"`
	SegmentBase          *SegmentBase     `xml:"SegmentBase,omitempty"`
	SegmentList          *SegmentList     `xml:"SegmentList,omitempty"`
	SegmentTemplate      *SegmentTemplate `xml:"SegmentTemplate,omitempty"`
	AssetIdentifier      *Descriptor      `xml:"AssetIdentifier,omitempty"`
	EventStream          []*EventStream   `xml:"EventStream,omitempty"`
	AdaptationSet        []*AdaptationSet `xml:"AdaptationSet,omitempty"`
	Subset               []*Subset        `xml:"Subset,omitempty"`
	SupplementalProperty []*Descriptor    `xml:"SupplementalProperty,omitempty"`

	Id                 string   `xml:"id,attr,omitempty"`
	Start              Duration `xml:"start,attr,omitempty"`
	Duration           Duration `xml:"duration,attr,omitempty"`
	BitstreamSwitching bool     `xml:"bitstreamSwitching,attr,omitempty"` // default: false
}

type EventStream struct {
	Event []*Event `xml:"Event,omitempty"`

	SchemeIdUri string `xml:"schemeIdUri,attr"`
	Value       string `xml:"value,attr,omitempty"`
	Timescale   uint   `xml:"timescale,attr,omitempty"`
}

type Event struct {
	PresentationTime uint64 `xml:"presentationTime,attr,omitempty"` // default: 0
	Duration         uint64 `xml:"duration,attr,omitempty"`
	Id               uint   `xml:"id,attr,omitempty"`
	MessageData      string `xml:"messageData,attr,omitempty"`
}

type AdaptationSet struct {
	*RepresentationBase

	Accessibility    []*Descriptor       `xml:"Accessibility,omitempty"`
	Role             []*Descriptor       `xml:"Role,omitempty"`
	Rating           []*Descriptor       `xml:"Rating,omitempty"`
	Viewpoint        []*Descriptor       `xml:"Viewpoint,omitempty"`
	ContentComponent []*ContentComponent `xml:"ContentComponent,omitempty"`
	BaseURL          []*BaseURL          `xml:"BaseURL,omitempty"`
	SegmentBase      *SegmentBase        `xml:"SegmentBase,omitempty"`
	SegmentList      *SegmentList        `xml:"SegmentList,omitempty"`
	SegmentTemplate  *SegmentTemplate    `xml:"SegmentTemplate,omitempty"`
	Representation   []*Representation   `xml:"Representation,omitempty"`

	Id                      uint            `xml:"id,attr,omitempty"`
	Group                   uint            `xml:"group,attr,omitempty"`
	Lang                    string          `xml:"lang,attr,omitempty"`
	ContentType             string          `xml:"contentType,attr,omitempty"`
	Par                     Ratio           `xml:"par,attr,omitempty"`
	MinBandwidth            uint            `xml:"minBandwidth,attr,omitempty"`
	MaxBandwidth            uint            `xml:"maxBandwidth,attr,omitempty"`
	MinWidth                uint            `xml:"minWidth,attr,omitempty"`
	MaxWidth                uint            `xml:"maxWidth,attr,omitempty"`
	MinHeight               uint            `xml:"minHeight,attr,omitempty"`
	MaxHeight               uint            `xml:"maxHeight,attr,omitempty"`
	MinFrameRate            FrameRate       `xml:"minFrameRate,attr,omitempty"`
	MaxFrameRate            FrameRate       `xml:"maxFrameRate,attr,omitempty"`
	SegmentAlignment        ConditionalUint `xml:"segmentAlignment,attr,omitempty"`        // default: false
	SubsegmentAlignment     ConditionalUint `xml:"subsegmentAlignment,attr,omitempty"`     // default: false
	SubsegmentStartsWithSAP SAP             `xml:"subsegmentStartsWithSAP,attr,omitempty"` // default: 0
	BitstreamSwitching      bool            `xml:"bitstreamSwitching,attr,omitempty"`
}

type ContentComponent struct {
	Accessibility []*Descriptor `xml:"Accessibility,omitempty"`
	Role          []*Descriptor `xml:"Role,omitempty"`
	Rating        []*Descriptor `xml:"Rating,omitempty"`
	Viewpoint     []*Descriptor `xml:"Viewpoint,omitempty"`

	Id          uint   `xml:"id,attr,omitempty"`
	Lang        string `xml:"lang,attr,omitempty"`
	ContentType string `xml:"contentType,attr,omitempty"`
	Par         Ratio  `xml:"par,attr,omitempty"`
}

type Representation struct {
	*RepresentationBase

	BaseURL           []*BaseURL           `xml:"BaseURL,omitempty"`
	SubRepresentation []*SubRepresentation `xml:"SubRepresentation,omitempty"`
	SegmentBase       *SegmentBase         `xml:"SegmentBase,omitempty"`
	SegmentList       *SegmentList         `xml:"SegmentList,omitempty"`
	SegmentTemplate   *SegmentTemplate     `xml:"SegmentTemplate,omitempty"`

	Id                     StringNoWhitespace `xml:"id,attr"`
	Bandwidth              uint               `xml:"bandwidth,attr"`
	QualityRanking         uint               `xml:"qualityRanking,attr,omitempty"`
	DependencyId           StringVector       `xml:"dependencyId,attr,omitempty"`
	MediaStreamStructureId StringVector       `xml:"mediaStreamStructureId,attr,omitempty"`
}

type SubRepresentation struct {
	*RepresentationBase

	Level            uint         `xml:"level,attr,omitempty"`
	DependencyLevel  UIntVector   `xml:"dependencyLevel,attr,omitempty"`
	Bandwidth        uint         `xml:"bandwidth,attr,omitempty"`
	ContentComponent StringVector `xml:"contentComponent,attr,omitempty"`
}

type RepresentationBase struct {
	FramePacking              []*Descriptor  `xml:"FramePacking,omitempty"`
	AudioChannelConfiguration []*Descriptor  `xml:"AudioChannelConfiguration,omitempty"`
	ContentProtection         []*Descriptor  `xml:"ContentProtection,omitempty"`
	EssentialProperty         []*Descriptor  `xml:"EssentialProperty,omitempty"`
	SupplementalProperty      []*Descriptor  `xml:"SupplementalProperty,omitempty"`
	InbandEventStream         []*EventStream `xml:"InbandEventStream,omitempty"`

	Profiles          string    `xml:"profiles,attr,omitempty"`
	Width             uint      `xml:"width,attr,omitempty"`
	Height            uint      `xml:"height,attr,omitempty"`
	Sar               Ratio     `xml:"sar,attr,omitempty"`
	FrameRate         FrameRate `xml:"frameRate,attr,omitempty"`
	AudioSamplingRate string    `xml:"audioSamplingRate,attr,omitempty"`
	MimeType          string    `xml:"mimeType,attr,omitempty"`
	SegmentProfiles   string    `xml:"segmentProfiles,attr,omitempty"`
	Codecs            string    `xml:"codecs,attr,omitempty"`
	MaximumSAPPeriod  float64   `xml:"maximumSAPPeriod,attr,omitempty"`
	StartWithSAP      SAP       `xml:"startWithSAP,attr,omitempty"`
	MaxPlayoutRate    float64   `xml:"maxPlayoutRate,attr,omitempty"`
	CodingDependency  bool      `xml:"codingDependency,attr,omitempty"`
	ScanType          VideoScan `xml:"scanType,attr,omitempty"`
}

type Subset struct {
	Contains UIntVector `xml:"contains,attr"`
	Id       string     `xml:"id,attr,omitempty"`
}

type SegmentBase struct {
	Initialization      *URL `xml:"Initialization,omitempty"`
	RepresentationIndex *URL `xml:"RepresentationIndex,omitempty"`

	Timescale                uint    `xml:"timescale,attr,omitempty"`
	PresentationTimeOffset   uint64  `xml:"presentationTimeOffset,attr,omitempty"`
	IndexRange               string  `xml:"indexRange,attr,omitempty"`
	IndexRangeExact          bool    `xml:"indexRangeExact,attr,omitempty"` // default: false
	AvailabilityTimeOffset   float64 `xml:"availabilityTimeOffset,attr,omitempty"`
	AvailabilityTimeComplete bool    `xml:"availabilityTimeComplete,attr,omitempty"`
}

type MultipleSegmentBase struct {
	*SegmentBase

	SegmentTimeline    *SegmentTimeline `xml:"SegmentTimeline,omitempty"`
	BitstreamSwitching *URL             `xml:"BitstreamSwitching,omitempty"`

	Duration    uint `xml:"duration,attr,omitempty"`
	StartNumber uint `xml:"startNumber,attr,omitempty"`
}

type URL struct {
	SourceURL string `xml:"sourceURL,attr,omitempty"`
	Range     string `xml:"range,attr,omitempty"`
}

type SegmentList struct {
	*MultipleSegmentBase

	SegmentURL []*SegmentURL `xml:"SegmentURL,omitempty"`
}

type SegmentURL struct {
	Media      string `xml:"media,attr,omitempty"`
	MediaRange string `xml:"mediaRange,attr,omitempty"`
	Index      string `xml:"index,attr,omitempty"`
	IndexRange string `xml:"indexRange,attr,omitempty"`
}

type SegmentTemplate struct {
	*MultipleSegmentBase

	Media              string `xml:"media,attr,omitempty"`
	Index              string `xml:"index,attr,omitempty"`
	Initialization     string `xml:"initialization,attr,omitempty"`
	BitstreamSwitching string `xml:"bitstreamSwitching,attr,omitempty"`
}

type SegmentTimeline struct {
	S []struct {
		T uint64 `xml:"t,attr,omitempty"`
		N uint64 `xml:"n,attr,omitempty"`
		D uint64 `xml:"d,attr"`
		R int    `xml:"r,attr,omitempty"` // default: 0
	}
}

type BaseURL struct {
	CDATA string `xml:",chardata"`

	ServiceLocation          string  `xml:"serviceLocation,attr,omitempty"`
	ByteRange                string  `xml:"byteRange,attr,omitempty"`
	AvailabilityTimeOffset   float64 `xml:"availabilityTimeOffset,attr,omitempty"`
	AvailabilityTimeComplete bool    `xml:"availabilityTimeComplete,attr,omitempty"`
}

type ProgramInformation struct {
	Title     string `xml:"Title,omitempty"`
	Source    string `xml:"Source,omitempty"`
	Copyright string `xml:"Copyright,omitempty"`

	Lang               string `xml:"lang,attr,omitempty"`
	MoreInformationURL string `xml:"moreInformationURL,attr,omitempty"`
}

type Descriptor struct {
	SchemeIdUri string `xml:"schemeIdUri,attr"`
	Value       string `xml:"value,attr,omitempty"`
	Id          string `xml:"id,attr,omitempty"`
}

type Metrics struct {
	Reporting []*Descriptor `xml:"Reporting"`
	Range     []*Range      `xml:"Range,omitempty"`

	Metrics string `xml:"metrics,attr"`
}

type Range struct {
	Starttime Duration `xml:"starttime,attr,omitempty"`
	Duration  Duration `xml:"duration,attr,omitempty"`
}
