package socketfiles

import (
	"net/url"
	"time"
)

type throttler struct {
	bitrate int

	delay time.Duration
	next  *time.Timer
}

func (t *throttler) updateDelay(prescale int) {
	if t.bitrate <= 0 {
		t.delay = 0
		t.next = nil
		return
	}

	// delay = nanoseconds per byte
	t.delay = (8 * time.Second) / time.Duration(t.bitrate)
	t.next = time.NewTimer(0)

	// recalculate to the actual expected maximum bitrate
	t.bitrate = int(8 * time.Second / t.delay)

	if prescale > 1 {
		t.delay *= time.Duration(prescale)
	}
}

func (t *throttler) throttle(scale int) {
	if t.next == nil {
		return
	}

	<-t.next.C

	if scale > 1 {
		t.next.Reset(time.Duration(scale) * t.delay)
		return
	}

	t.next.Reset(t.delay)
}

func (t *throttler) set(q url.Values) error {
	if bitrate, ok, err := getSize(q, FieldMaxBitrate); ok || err != nil {
		if err != nil {
			return err
		}

		t.bitrate = bitrate
	}

	return nil
}
