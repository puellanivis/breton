package metrics

import (
	"time"
)

type Timer struct {
	t time.Time
	o Observer
}

func (t *Timer) Done() {
	if t.o == nil {
		return
	}

	t.o.Observe(time.Since(t.t).Seconds())
}

func newTimer(o Observer) *Timer {
	return &Timer{
		t: time.Now(),
		o: o,
	}
}
