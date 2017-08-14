// Package dash provides an abstraction to accessing DASH streams.
package dash

import (
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

	metrics *metricsPack

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

	url := urlLabel.WithValue(manifest)

	idx := strings.LastIndexByte(manifest, '/')
	base, manifest := manifest[:idx+1], manifest[idx+1:]

	adaptations := make(map[string]*adaptation)

	for _, p := range m.Period {
		pid := p.Id

		for _, as := range p.AdaptationSet {
			a := &adaptation{
				pid: pid,
				aid: as.Id,
			}

			if tmpl := as.SegmentTemplate; tmpl != nil {
				a.init = base + tmpl.Initialization
				a.media = base + tmpl.Media
			}

			for _, r := range as.Representation {
				if as.MimeType == "" {
					as.MimeType = r.MimeType
				}

				a.reps = append(a.reps, r)
			}

			if as.MimeType == "" {
				continue
			}

			adaptations[as.MimeType] = a
		}
	}

	minTime := m.MinimumUpdatePeriod.Duration
	// I can’t imagine using a minimum time of less than one millisecond.
	if minTime < time.Millisecond {
		minTime = time.Millisecond
	}

	return &Manifest{
		base:        base,
		manifest:    manifest,
		dynamic:     m.Type == "dynamic",
		adaptations: adaptations,
		m:           newMPD(manifest, minTime),

		metrics: baseMetrics.WithLabels(url),
	}, nil
}

// MinimumUpdatePeriod returns the shortest period within with a Manifest’s
// information is to update.
func (m *Manifest) MinimumUpdatePeriod() time.Duration {
	return m.m.expTime
}

// Stream returns a Stream object that specifies a specific series of segments within the DASH manifest.
func (m *Manifest) Stream(w io.Writer, mimeType string, picker Picker) (*Stream, error) {
	if picker == nil {
		picker = PickFirst()
	}

	as, ok := m.adaptations[mimeType]
	if !ok {
		return nil, errors.New("mime-type not available")
	}

	var picked *mpd.Representation

	for _, rep := range as.reps {
		// I don’t know how this could end up being here,
		// but let’s discard it regardless.
		if rep == nil {
			continue
		}

		picked = picker(rep)
	}

	if picked == nil {
		return nil, errors.New("no suitable representations found")
	}

	if glog.V(1) {
		glog.Infof("chose %s with bandwidth: %d", mimeType, picked.Bandwidth)
		if picked.Height > 0 && picked.Width > 0 {
			glog.Infof("chose %s with %d×%d", mimeType, picked.Height, picked.Width)
		}
	}

	return &Stream{
		w: w,
		m: m,

		metrics: m.metrics.WithLabels(typeLabel.WithValue(mimeType)),

		pid: as.pid,
		aid: as.aid,

		dynamic: m.dynamic,
		init:    as.init,
		media:   as.media,

		bw:    picked.Bandwidth,
		repID: string(picked.Id),
	}, nil
}
