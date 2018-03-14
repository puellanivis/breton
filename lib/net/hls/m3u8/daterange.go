package m3u8

import (
	"time"
)

type DateRange struct {
	ID    string `m3u8:"ID"`
	Class string `m3u8:"CLASS,optional"`

	StartDate time.Time `m3u8:"START-DATE" format:"2006-01-02T15:04:05.999Z07:00"`
	EndDate   time.Time `m3u8:"END-DATE,optional" format:"2006-01-02T15:04:05.999Z07:00"`

	Duration        time.Duration `m3u8:"DURATION,optional"`
	PlannedDuration time.Duration `m3u8:"PLANNED-DURATION,optional"`

	// SCTE35-CMD
	// SCTE35-OUT
	// SCTE35-IN

	EndOnNext bool `m3u8:"END-ON-NEXT,optional" enum:",YES"`

	ClientAttribute map[string]interface{} `m3u8:"X-,optional"`
}
