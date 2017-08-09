package mpd

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Duration struct {
	time.Duration
}

var errInvalidDur = errors.New("invalid duration")

const (
	// These values are so naive
	Day   = 24 * time.Hour
	Month = Day * 30
	Year  = Day * 365
)

func (d Duration) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if s := d.XMLString(); s != "" {
		return xml.Attr{
			Name:  name,
			Value: s,
		}, nil
	}

	return xml.Attr{}, nil
}

func (d Duration) XMLString() string {
	seconds := int(d.Duration.Seconds())

	if seconds == 0 {
		return ""
	}

	var neg bool
	if seconds < 0 {
		neg = true
		seconds = -seconds
	}

	minutes, seconds := seconds/60, seconds%60
	hours, minutes := minutes/60, minutes%60
	days, hours := hours/24, hours%24
	years, days := days/365, days%365
	months, days := days/30, days%30

	var elems []string
	if neg {
		elems = append(elems, "-")
	}
	elems = append(elems, "P")

	if years > 0 {
		elems = append(elems, fmt.Sprintf("%dY", years))
	}
	if months > 0 {
		elems = append(elems, fmt.Sprintf("%dM", months))
	}
	if days > 0 {
		elems = append(elems, fmt.Sprintf("%dD", days))
	}
	if hours > 0 || minutes > 0 || seconds > 0 {
		elems = append(elems, "T")

		if hours > 0 {
			elems = append(elems, fmt.Sprintf("%dH", hours))
		}
		if minutes > 0 {
			elems = append(elems, fmt.Sprintf("%dM", minutes))
		}
		if seconds > 0 {
			elems = append(elems, fmt.Sprintf("%dS", seconds))
		}
	}

	return strings.Join(elems, "")
}

func (d *Duration) UnmarshalXMLAttr(attr xml.Attr) error {
	if len(attr.Value) < 1 {
		return errInvalidDur
	}

	var neg bool
	if attr.Value[0] == '-' {
		neg = true
		attr.Value = attr.Value[1:]
	}

	if len(attr.Value) < 1 || attr.Value[0] != 'P' {
		return errInvalidDur
	}

	var radix bool
	var scale time.Duration = 1
	var n time.Duration
	var t bool

	for _, r := range attr.Value[1:] {
		switch {
		case r >= '0' && r <= '9':
			n *= 10
			n += time.Duration(r - '0')
			if radix {
				scale *= 10
			}

		case r == 'T' && !t:
			t = true

		case r == 'Y' && !t:
			d.Duration += n * Year
			n = 0
		case r == 'M' && !t:
			d.Duration += n * Month
		case r == 'D' && !t:
			d.Duration += n * Day

		case r == '.' && t && !radix:
			radix = true

		case r == 'H' && t:
			d.Duration += n * time.Hour
			n = 0
		case r == 'M' && t:
			d.Duration += n * time.Minute
			n = 0
		case r == 'S' && t:
			d.Duration += (n * time.Second) / scale
			n = 0

		default:
			return errInvalidDur
		}
	}

	if neg {
		d.Duration = -d.Duration
	}

	return nil
}
