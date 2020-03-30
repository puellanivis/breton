package socketfiles

import (
	"time"
)

type throttler struct {
	bitrate int

	delay time.Duration
	next  *time.Timer
}

func (t *throttler) drain() {
	if t.next == nil {
		return
	}

	if !t.next.Stop() {
		<-t.next.C
	}
}

func (t *throttler) updateDelay(prescale int) {
	if t.bitrate <= 0 {
		t.delay = 0
		t.drain()
		t.next = nil
		return
	}

	if t.next != nil {
		t.drain()
		t.next.Reset(0)
	} else {
		t.next = time.NewTimer(0)
	}

	// delay = nanoseconds per byte
	t.delay = (8 * time.Second) / time.Duration(t.bitrate)

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

func (t *throttler) setBitrate(bitrate, prescale int) int {
	prev := t.bitrate

	t.bitrate = bitrate
	t.updateDelay(prescale)

	return prev
}
