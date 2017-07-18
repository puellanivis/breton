package dash

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"lib/files"
	"lib/log"
	"lib/net/dash/mpd"
	//"github.com/zencoder/go-dash/mpd"
)

type adaptation struct {
	// indexes into the period and adaptions
	pid string
	aid uint

	// template strings
	init, media string
	startNum    uint

	reps []*mpd.Representation
}

// A Manifest holds the essential identifying information about a DASH manifest it is used to generate Streams.
type Manifest struct {
	base     string
	manifest string

	dynamic     bool
	adaptations map[string]*adaptation

	m *cachedMPD
}

// New returns a Manifest constructed from the given URL.
func New(ctx context.Context, manifest string) (*Manifest, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	m, err := readMPD(ctx, manifest)
	if err != nil {
		return nil, err
	}

	idx := strings.LastIndexByte(manifest, '/')
	base, manifest := manifest[:idx+1], manifest[idx+1:]

	adaptations := make(map[string]*adaptation)

	for _, p := range m.Period {
		pid := p.Id

		for _, as := range p.AdaptationSet {
			if as.MimeType == "" {
				continue
			}

			a := &adaptation{
				pid: pid,
				aid: as.Id,
			}

			if tmpl := as.SegmentTemplate; tmpl != nil {
				a.init = base + tmpl.Initialization
				a.media = base + tmpl.Media
			}

			for _, r := range as.Representation {
				a.reps = append(a.reps, r)
			}

			adaptations[as.MimeType] = a
		}
	}

	// the xs.duration type has a minimum resolution of PT1S (1 second)
	// it’s possible they could implement PT1m, or PT1n, but not now…
	minTime := m.MinimumUpdatePeriod.Duration
	if minTime < 1*time.Second {
		minTime = 1 * time.Second
	}

	return &Manifest{
		base:        base,
		manifest:    manifest,
		dynamic:     m.Type == "dynamic",
		adaptations: adaptations,
		m:           newMPD(manifest, minTime),
	}, nil
}

// MinimumUpdatePeriod returns the shortest period within with a Manifest’s
// information is to update.
func (m *Manifest) MinimumUpdatePeriod() time.Duration {
	return m.m.expTime
}

// Pull reads the next series of segments in the Stream. If time.Duration
// is 0, it means that no new segments were available. (You’re likely calling
// this function too often in this case. Calling any faster than the duration
// returned by MinimumUpdatePeriod is just a waste of cycles.)
func (m *Manifest) Pull(ctx context.Context, s *Stream) (time.Duration, error) {
	ctx, err := files.WithRoot(ctx, m.base)
	if err != nil {
		return 0, err
	}

	cur, err := m.m.get(ctx)
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

// Stream returns a Stream object that specifies a specific series of segments within the DASH manifest.
func (m *Manifest) Stream(w io.Writer, mimeType string, picker PickRepFunc) (*Stream, error) {
	if picker == nil {
		picker = PickFirst
	}

	as, ok := m.adaptations[mimeType]
	if !ok {
		return nil, errors.New("mime-type not available")
	}

	var best *mpd.Representation

	for _, rep := range as.reps {
		// I don’t know how this could end up being here,
		// but let’s discard it regardless.
		if rep == nil {
			continue
		}

		if picker(best, rep) {
			best = rep
		}
	}

	if best == nil {
		return nil, errors.New("no suitable representations found")
	}

	if log.V(1) {
		log.Infof("chose %s with bandwidth: %d", mimeType, best.Bandwidth)
		if best.Height > 0 && best.Width > 0 {
			log.Infof("chose %s with %d×%d", mimeType, best.Height, best.Width)
		}
	}

	return &Stream{
		w: w,

		pid: as.pid,
		aid: as.aid,

		dynamic: m.dynamic,
		init:    as.init,
		media:   as.media,

		Bandwidth: best.Bandwidth,
		RepID:     string(best.Id),
	}, nil
}
