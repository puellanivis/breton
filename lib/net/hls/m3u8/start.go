package m3u8

import (
	"time"
)

// Start implements the START directive of the m3u8 standard.
type Start struct {
	TimeOffset time.Duration `m3u8:"TIME-OFFSET"`
	Precise    bool          `m3u8:"PRECISE,optional" enum:"NO,YES"`
}
