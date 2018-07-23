package mpd

import (
	"encoding/xml"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// Duration naively implements the xsd:duration format defined by https://www.w3.org/TR/2004/REC-xmlschema-2-20041028/datatypes.html#duration
// N.B.: It is not intended to serve as a general xsd:duration as it does not implement the month side of xsd:duration.
type Duration struct {
	ns time.Duration
	m  int
}

// A copy of various time.Duration constants.
const (
	Day         = 24 * time.Hour
	Hour        = time.Hour
	Minute      = time.Minute
	Second      = time.Second
	Millisecond = time.Millisecond
	Microsecond = time.Microsecond
	Nanosecond  = time.Nanosecond
)

// Duration returns the mpd.Duration value as a time.Duration.
//
// If Duration contains a non-zero value for the month side of the xsd:duration,
// then this function will return an error noting that the value is out of range..
func (d Duration) Duration() (time.Duration, error) {
	if d.m != 0 {
		return 0, errors.New("value out of range")
	}

	return d.ns, nil
}

// Add returns the mpd.Duration plus the given time.Duration.
func (d Duration) Add(dur time.Duration) Duration {
	d.ns += dur

	return d
}

// Scale returns the mpd.Duration scaled by the given time.Duration.
func (d Duration) Scale(dur time.Duration) Duration {
	d.ns *= dur

	return d
}

// AddToTime returns the given time.Time plus the given mpd.Duration value..
func (d Duration) AddToTime(t time.Time) time.Time {
	if d.m != 0 {
		t = t.AddDate(0, d.m, 0)
	}

	return t.Add(d.ns)
}

// MarshalXMLAttr implements xml.MarshalerAttr.
func (d Duration) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if s := d.XMLString(); s != "" {
		return xml.Attr{
			Name:  name,
			Value: s,
		}, nil
	}

	return xml.Attr{}, nil
}

// NewDuration returns an mpd.Duration corresponding to the length in months and time.Duration.
//
// It returns an error if the sign of two arguments are different.
func NewDuration(months int, d time.Duration) (Duration, error) {
	if (d < 0 && months > 0) || (d > 0 && months < 0) {
		return Duration{}, errors.New("invalid duration")
	}

	return Duration{
		ns: d,
		m:  months,
	}, nil
}

// XMLString returns the XML representation of the Duration as a string.
func (d Duration) XMLString() string {
	var b []byte
	dur := d.ns
	m := int64(d.m)

	if dur < 0 {
		dur = -dur
		m = -m
		b = append(b, '-')
	}

	b = append(b, 'P')

	seconds, frac := int64(dur/Second), int64(dur%Second)

	minutes, seconds := seconds/60, seconds%60
	hours, minutes := minutes/60, minutes%60
	days, hours := hours/24, hours%24

	years, months := m/12, m%12

	if years > 0 {
		b = strconv.AppendInt(b, years, 10)
		b = append(b, 'Y')
	}

	if months > 0 {
		b = strconv.AppendInt(b, months, 10)
		b = append(b, 'M')
	}

	if days > 0 {
		b = strconv.AppendInt(b, days, 10)
		b = append(b, 'D')
	}

	if hours > 0 || minutes > 0 || seconds > 0 || frac > 0 {
		b = append(b, 'T')

		if hours > 0 {
			b = strconv.AppendInt(b, hours, 10)
			b = append(b, 'H')
		}
		if minutes > 0 {
			b = strconv.AppendInt(b, minutes, 10)
			b = append(b, 'M')
		}

		if seconds+frac != 0 || hours+minutes == 0 {
			b = strconv.AppendInt(b, seconds, 10)
			if frac > 0 {
				b = append(b, '.')

				f := strconv.AppendInt(nil, frac, 10)

				if len(f) < 9 {
					nb := make([]byte, 9-len(f))
					nb[0] = '0'

					bp := 1
					for bp < len(nb) {
						copy(nb[bp:], nb[:bp])
						bp *= 2
					}

					b = append(b, nb...)
				}

				l := len(f)
				for f[l-1] == '0' {
					l--
				}

				b = append(b, f[:l]...)
			}
			b = append(b, 'S')
		}
	}

	return string(b)
}

var durationScales = []time.Duration{
	Second,

	100 * Millisecond,
	10 * Millisecond,
	Millisecond,

	100 * Microsecond,
	10 * Microsecond,
	Microsecond,

	100 * Nanosecond,
	10 * Nanosecond,
	Nanosecond,
}

// UnmarshalXMLAttr implements xml.UnmarshalerAttr.
func (d *Duration) UnmarshalXMLAttr(attr xml.Attr) error {
	var dur time.Duration
	var months int

	v := attr.Value

	if len(v) < 1 {
		return errors.New("invalid duration")
	}

	var neg bool
	if v[0] == '-' {
		neg = true
		v = v[1:]
	}

	if len(v) < 1 || v[0] != 'P' {
		return errors.New("invalid duration")
	}

	var n time.Duration

	var radix bool
	var scale uint

	var t, end bool
	for _, r := range v[1:] {
		if end {
			// we had extra chars after a fractional value, this is an error
			return errors.New("invalid duration")
		}

		switch r {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			n = (n * 10) + time.Duration(r-'0')
			if radix {
				scale++
			}

		case 'T':
			if t {
				return errors.New("invalid duration")
			}
			t = true

		case '.':
			if radix {
				return errors.New("invalid duration")
			}
			radix = true

		case 'Y':
			if t || radix {
				return errors.New("invalid duration")
			}

			months += int(n) * 12
			n = 0

		case 'D':
			if t || radix {
				return errors.New("invalid duration")
			}

			dur += n * Day
			n = 0

		case 'H':
			if !t || radix {
				return errors.New("invalid duration")
			}

			dur += n * Hour
			n = 0

		case 'M':
			switch {
			case radix:
				return errors.New("invalid duration")

			case t:
				dur += n * Minute

			default:
				months += int(n)
			}
			n = 0

		case 'S':
			if !t || scale > 9 {
				return errors.New("invalid duration")
			}

			dur += n * durationScales[scale]
			end = true

		default:
			return errors.New("invalid duration")
		}
	}

	if neg {
		dur = -dur
	}

	d.ns = dur
	d.m = months

	return nil
}
