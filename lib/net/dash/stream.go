package dash

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/glog"
	"github.com/puellanivis/breton/lib/net/dash/mpd"
)

// Stream is a structure holding the information necessary to retreive a#
// DASH stream.
type Stream struct {
	w io.Writer
	m *Manifest

	metrics *metricsPack

	// so we can find the appropriate SegmentTimeline
	// in a given mpd.MPD
	pid string
	aid uint

	eof     bool
	dynamic bool

	init, media string
	bw          uint
	repID       string
	time        uint64
}

// Bandwidth returns the bandwidth that was selected for this Stream.
func (s *Stream) Bandwidth() uint {
	return s.bw
}

// RepresentationID returns the ID of the Representationn that was selected for this Stream.
func (s *Stream) RepresentationID() string {
	return s.repID
}

// buildURL takes the given template, and given number, and renders it
// according to the DASH standards.
func (s *Stream) buildURL(template string, number uint) string {
	b := new(bytes.Buffer)

	tmpl := []byte(template)

	for {
		i := bytes.IndexByte(tmpl, '$')
		if i < 0 {
			break
		}

		b.Write(tmpl[:i])
		tmpl = tmpl[i+1:]

		j := bytes.IndexByte(tmpl, '$')
		if j < 0 {
			break
		}

		var v string
		v, tmpl = string(tmpl[:j]), tmpl[j+1:]

		format, val := "%v", v

		idx := strings.IndexByte(v, '%')
		if idx > 0 {
			format, val = v[idx:], v[:idx]
		}

		switch val {
		case "":
			fmt.Fprint(b, "$")

		case "Time":
			fmt.Fprintf(b, format, s.time)
		case "Bandwidth":
			fmt.Fprintf(b, format, s.bw)
		case "RepresentationID":
			fmt.Fprintf(b, format, s.repID)
		case "Number":
			fmt.Fprintf(b, format, number)
		}
	}

	b.Write(tmpl)

	return b.String()
}

// Init reads the initialization URL from the stream.
func (s *Stream) Init(ctx context.Context) error {
	init := s.buildURL(s.init, 0)
	return s.readFrom(ctx, init, 0)
}

// readTo reads a given URL into the Stream’s io.Writer while keeping metrics.
func (s *Stream) readFrom(ctx context.Context, url string, scale float64) error {
	done := s.metrics.timing.Timer()
	defer done()

	if glog.V(5) {
		glog.Info("Grabbing:", url)
	}

	f, err := files.Open(ctx, url)
	if err != nil {
		return err
	}
	defer f.Close()

	n, err := io.Copy(s.w, f)
	if scale > 0.1 {
		// n is measured in bytes, we want to record in bits per time unit
		s.metrics.bandwidth.Observe(float64(n*8) * scale)
	}

	return err
}

func (s *Stream) readTimeline(ctx context.Context, ts uint, num uint, tl *mpd.SegmentTimeline) (time.Duration, error) {
	if s.eof {
		return 0, io.EOF
	}

	var dur, tdur time.Duration
	var tscale = time.Second
	if ts != 0 {
		tscale = time.Second / time.Duration(ts)
	}

	for _, seg := range tl.S {
		t := seg.T

		if s.time <= t {
			s.time = t
		}

		dur = time.Duration(seg.D) * tscale
		scale := 1 / dur.Seconds()

		// less than or equal to, because it includes an implicit first
		for i := 0; i <= seg.R; i++ {
			if t <= s.time {
				t += seg.D
				continue
			}

			s.time = t

			url := s.buildURL(s.media, num)

			if err := s.readFrom(ctx, url, scale); err != nil {
				return tdur, err
			}

			num++
			tdur += dur
			t += seg.D
		}
	}

	if !s.dynamic {
		s.eof = true
		return tdur, io.EOF
	}

	return tdur, nil
}

// Read reads the next series of segments in the Stream.
// If the returned time.Duration is 0, it means that no new segments were available.
// (In which case, you’re likely calling this function too often in this case.
// Calling any faster than the duration returned by MinimumUpdatePeriod is just a waste of cycles.)
func (s *Stream) Read(ctx context.Context) (time.Duration, error) {
	ctx, err := files.WithRoot(ctx, s.m.base)
	if err != nil {
		return 0, err
	}

	cur, err := s.m.m.get(ctx)
	if err != nil {
		return 0, err
	}

	for _, p := range cur.Period {
		if p.Id != s.pid {
			continue
		}

		for _, as := range p.AdaptationSet {
			if as.Id != s.aid {
				continue
			}

			tmpl := as.SegmentTemplate
			if tmpl == nil {
				return 0, errors.New("SegmentTemplate not found")
			}

			if tmpl.SegmentTimeline == nil {
				return 0, errors.New("SegmentTimeline not found")
			}

			return s.readTimeline(ctx, tmpl.Timescale, tmpl.StartNumber, tmpl.SegmentTimeline)
		}
	}

	return 0, fmt.Errorf("no media could be found for pid: %s, aid: %d", s.pid, s.aid)
}
