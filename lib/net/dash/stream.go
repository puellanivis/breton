package dash

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"lib/files"
	"lib/log"
	"lib/metrics"
	"lib/net/dash/mpd"
	//"lib/util"
	//"github.com/zencoder/go-dash/mpd"
)

var (
	sizes   = metrics.Summary("dash_segment_sizes_bps", "tracks the bits per second of dash segments received", metrics.WithObjective(0.5, 0.05), metrics.WithObjective(0.9, 0.01), metrics.WithObjective(0.99, 0.001))
	timings = metrics.Summary("dash_segment_timings_seconds", "tracks how long it takes to receive segments", metrics.WithObjective(0.5, 0.05), metrics.WithObjective(0.9, 0.01), metrics.WithObjective(0.99, 0.001))
)

// Stream is a structure holding the information necessary to retreive a#
// DASH stream.
type Stream struct {
	w io.Writer

	// so we can find the appropriate SegmentTimeline
	// in a given mpd.MPD
	pid string
	aid uint

	eof bool

	dynamic     bool
	init, media string

	Bandwidth uint
	RepID     string
	time      uint64
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
			fmt.Fprintf(b, format, s.Bandwidth)
		case "RepresentationID":
			fmt.Fprintf(b, format, s.RepID)
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
	timer := timings.Timer()
	defer timer.Done()

	if log.V(5) {
		log.Info("Grabbing:", url)
	}

	n, err := files.ReadTo(ctx, s.w, url)
	if scale > 0.1 {
		// n is measured in bytes, we want to record in bits
		sizes.Observe(float64(n*8) * scale)
	}

	return err
}

func (s *Stream) readTimeline(ctx context.Context, ts uint, num uint, tl *mpd.SegmentTimeline) (time.Duration, error) {
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

		//tdur = (dur * time.Duration(r+1))
	}

	if !s.dynamic {
		s.eof = true
		return tdur, io.EOF
	}

	return tdur, nil
}

/*
// ReadSegments will read the currently available DASH segments into the
// Stream’s io.Writer. It returns a total time.Duration of segments read.
// If this is not a dynamic DASH stream, then it will return io.EOF when
// it reaches the end of the stream. It is not an error for a dynamic
// stream to return a 0 amount of time.Duration, it means that it has
// already read all of the available Segments already.
func (s *Stream) ReadSegments(ctx context.Context) (time.Duration, error) {
	ctx, err := files.WithRoot(ctx, s.base)
	if err != nil {
		return 0, err
	}

	m, err := s.getMPD(ctx)
	if err != nil {
		return 0, err
	}

	tmpl := m.Period.AdaptationSets[s.index].SegmentTemplate

	var dur time.Duration
	var cnt int
	var tscale = time.Second
	if tmpl.StartNumber != nil {
		cnt = int(*tmpl.StartNumber)
	}
	if tmpl.Timescale != nil {
		tscale = time.Second / time.Duration(*tmpl.Timescale)
	}
	media := *tmpl.Media

	var tdur time.Duration

	for _, seg := range tmpl.SegmentTimeline.Segments {
		if seg.StartTime == nil {
			continue
		}

		if s.time < *seg.StartTime {
			s.time = *seg.StartTime
		}

		dur = time.Duration(seg.Duration) * tscale
		scale := 1 / dur.Seconds()

		var r uint64
		if seg.RepeatCount != nil && *seg.RepeatCount > 0 {
			r = uint64(*seg.RepeatCount)
		}

		t := *seg.StartTime

		// less than or equal to, because it includes an implicit first
		for i := uint64(0); i <= r; i++ {
			if t <= s.time {
				t += seg.Duration
				continue
			}

			cnt++
			s.time = t

			url := s.buildURL(media, cnt)

			if err := s.readFrom(ctx, url, scale); err != nil {
				return tdur, err
			}

			tdur += dur
			t += seg.Duration
		}

		//tdur = (dur * time.Duration(r+1))
	}

	if !s.dynamic {
		return tdur, io.EOF
	}

	return tdur, nil
}
// */
