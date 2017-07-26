package dash

import (
	"github.com/puellanivis/breton/lib/metrics"
)

const (
	urlLabel  = metrics.Label("manifest_url")
	typeLabel = metrics.Label("mime_type")
)

var (
	labels = metrics.WithLabels(urlLabel, typeLabel)
)

type metricsPack struct {
	timing    *metrics.SummaryValue
	bandwidth *metrics.SummaryValue
}

var baseMetrics = &metricsPack{
	timing:    metrics.Summary("dash_segment_timing_seconds", "tracks how long it takes to receive segments", labels, metrics.CommonObjectives()),
	bandwidth: metrics.Summary("dash_segment_bandwidth_bps", "tracks the bits per second of dash segments received", labels, metrics.CommonObjectives()),
}

func (m *metricsPack) WithLabels(labels ...metrics.Labeler) *metricsPack {
	return &metricsPack{
		timing:    m.timing.WithLabels(labels...),
		bandwidth: m.bandwidth.WithLabels(labels...),
	}
}
