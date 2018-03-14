package m3u8

import (
	"time"
)

type Start struct {
	TimeOffset time.Duration `m3u8:"TIME-OFFSET"`
	Precise    bool          `m3u8:"PRECISE,optional" enum:"NO,YES"`
}
