package mpd

import (
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// Duration implements the xsd:duration format defined by https://www.w3.org/TR/2004/REC-xmlschema-2-20041028/datatypes.html#duration
type Duration struct {
	time.Duration
}

var errInvalidDur = errors.New("invalid duration")

const (
	// These values are very _very_ naive
	Day   = 24 * time.Hour
	Month = Day * 30
	Year  = Day * 365
)

// MarshalXMLAttr implements the xml attribute marshaller interface.
func (d Duration) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if s := d.XMLString(); s != "" {
		return xml.Attr{
			Name:  name,
			Value: s,
		}, nil
	}

	return xml.Attr{}, nil
}

// XMLString returns the XML representation of the Duration as a string.
func (d Duration) XMLString() string {
	frac := d.Duration.Seconds()
	frac = frac - math.Floor(frac)

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

	frac += float64(seconds)

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
		if frac > 0 {
			if frac-float64(seconds) < 0.0005 {
				elems = append(elems, fmt.Sprintf("%dS", seconds))
			} else {
				elems = append(elems, fmt.Sprintf("%0.3fS", frac))
			}
		}
	}

	return strings.Join(elems, "")
}

// UnmarshalXMLAttr implements the xml attribute unmarshaller interface.
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
	var t, end bool

	for _, r := range attr.Value[1:] {
		if end {
			// we had extra chars after a fractional value, this is an error
			return errInvalidDur
		}

		switch {
		case r >= '0' && r <= '9':
			n *= 10
			n += time.Duration(r - '0')
			if radix {
				scale *= 10
			}
			continue
		case r == 'T' && !t:
			t = true
			continue
		case (r == '.' || r == ',') && !radix:
			radix = true
			continue

		case r == 'Y' && !t:
			d.Duration += n * Year
			n = 0
		case r == 'M' && !t:
			d.Duration += n * Month
		case r == 'D' && !t:
			d.Duration += n * Day

		case r == 'H' && t:
			d.Duration += n * time.Hour
			n = 0
		case r == 'M' && t:
			d.Duration += n * time.Minute
			n = 0
		case r == 'S' && t:
			d.Duration += n * time.Second
			n = 0

		default:
			return errInvalidDur
		}

		if radix {
			d.Duration /= scale
			end = true
		}
	}

	if neg {
		d.Duration = -d.Duration
	}

	return nil
}
