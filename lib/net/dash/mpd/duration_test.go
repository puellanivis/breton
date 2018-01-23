package mpd

import (
	"encoding/xml"
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	durs := map[time.Duration]string{
		time.Second:                        "PT1S",
		time.Minute:                        "PT1M",
		time.Hour:                          "PT1H",
		time.Second + time.Nanosecond:      "PT1.000000001S",
		time.Second + 10*time.Nanosecond:   "PT1.00000001S",
		time.Second + 100*time.Nanosecond:  "PT1.0000001S",
		time.Second + 1*time.Microsecond:   "PT1.000001S",
		time.Second + 10*time.Microsecond:  "PT1.00001S",
		time.Second + 100*time.Microsecond: "PT1.0001S",

		time.Second + 500*time.Microsecond:    "PT1.0005S",
		time.Second + time.Millisecond:        "PT1.001S",
		time.Second + 10*time.Millisecond:     "PT1.01S",
		time.Second + 100*time.Millisecond:    "PT1.1S",
		time.Hour + time.Second:               "PT1H1S",
		time.Minute + time.Second:             "PT1M1S",
		time.Hour + time.Minute:               "PT1H1M",
		time.Hour + time.Minute + time.Second: "PT1H1M1S",

		Day: "P1D",

		Day + time.Hour + time.Minute + time.Second: "P1DT1H1M1S",
	}

	var attr xml.Attr
	for dur, expect := range durs {
		d := Duration{
			ns: dur,
		}

		if d.XMLString() != expect {
			t.Errorf("Duration %s marshaled wrong: expected %q, got %q", dur, expect, d.XMLString())
		}

		var d2 Duration
		attr.Value = expect

		if err := d2.UnmarshalXMLAttr(attr); err != nil {
			t.Errorf("Error while Unmarshalling %q: %v", expect, err)
			continue
		}

		if d2.ns != dur {
			t.Errorf("XML value %q unmarshaled wrong: expected %s, got %s", expect, dur, d2.ns)
		}
	}
}
