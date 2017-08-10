package mpd

import (
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	durs := map[time.Duration]string{
		time.Second: "PT1S",
		time.Minute: "PT1M",
		time.Hour: "PT1H",
		time.Second + time.Nanosecond: "PT1S",
		time.Second + 10 * time.Nanosecond: "PT1S",
		time.Second + 100 * time.Nanosecond: "PT1S",
		time.Second + 1* time.Microsecond: "PT1S",
		time.Second + 10 * time.Microsecond: "PT1S",
		time.Second + 100 * time.Microsecond: "PT1S",
		// 1000500 µs is undefined, due to float-point rounding issues
		// but 1000501 µs is defined, and rounds up to 1001 ms
		time.Second + 501 * time.Microsecond: "PT1.001S",
		time.Second + time.Millisecond: "PT1.001S",
		time.Second + 10 * time.Millisecond: "PT1.010S",
		time.Second + 100 * time.Millisecond: "PT1.100S",
		time.Hour + time.Second: "PT1H1S",
		time.Minute + time.Second: "PT1M1S",
		time.Hour + time.Minute: "PT1H1M",
		time.Hour + time.Minute + time.Second: "PT1H1M1S",
		Year: "P1Y",
		Month: "P1M",
		Day: "P1D",
		Year + Month + Day + time.Hour + time.Minute + time.Second: "P1Y1M1DT1H1M1S",
	}

	for dur, expect := range durs {
		d := Duration{dur}

		if d.XMLString() != expect {
			t.Errorf("Duration %s marshaled wrong: expected %s, got %s", dur, expect, d.XMLString())
		}
	}
}
