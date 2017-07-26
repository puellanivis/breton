package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// A HistogramValue holds the tracking information for a specific Histogram or a “Child” of a Histogram.
type HistogramValue struct {
	// we want to duplicate this every WithLabels() call,
	// so we don’t use a pointer here.
	metric

	h  prometheus.Histogram
	hv *prometheus.HistogramVec
}

// WithLabels provides access to a labeled dimension of the metric, and returns a “Child” wherein the given labels are set.
// The “Child” returned is cacheable by the Caller, so as to avoid having to look it up again—this matters in latency-critical code.
func (h HistogramValue) WithLabels(labels ...Labeler) *HistogramValue {
	// we are working with a new copy, so no mutex is necessary.
	h.h = nil

	h.labels = h.labels.With(labels...)

	return &h
}

// Remove will remove a “Child” that matches the given labels from the metric, no longer exporting it.
func (h *HistogramValue) Remove(labels ...Labeler) bool {
	if h.hv == nil {
		return false
	}

	return h.hv.Delete(h.labels.getMap())
}

// Clear removes all “Children” from the metric.
func (h *HistogramValue) Clear() {
	if h.hv != nil {
		h.hv.Reset()
	}
}

// Histogram allows aggregated distributions of events, such as request latencies.
// This is at its core a Counter per bucket.
func Histogram(name string, help string, options ...Option) *HistogramValue {
	m := newMetric(name, help)

	m.histogramSettings = new(histogramSettings)

	for _, opt := range options {
		// in initialization, we throw the reverting function away
		_ = opt(m)
	}

	h := &HistogramValue{
		metric: *m,
	}

	opts := prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: m.buckets,
	}

	if h.labels != nil {
		if _, ok := m.labels.set.canSet["le"]; ok {
			panic("histograms cannot allow \"le\" as a label")
		}

		h.hv = prometheus.NewHistogramVec(opts, h.labels.set.keys)
		h.registry.MustRegister(h.hv)

	} else {
		h.h = prometheus.NewHistogram(opts)
		h.registry.MustRegister(h.h)
	}

	return h
}

// Observe records the given value into the Histogram.
func (h *HistogramValue) Observe(v float64) {
	if h.h == nil {
		// function is idempotent, and won’t step on others’ toes
		h.h = h.hv.With(h.labels.getMap()).(prometheus.Histogram)
	}

	h.h.Observe(v)
}

// ObserveDuration records the given Duration into the Histogram.
func (h *HistogramValue) ObserveDuration(d time.Duration) {
	h.Observe(d.Seconds())
}

// Timer times a piece of code and records to the Histogram its duration in seconds.
//
// (Caller MUST ensure the returned done function is called, and SHOULD use defer.)
func (h *HistogramValue) Timer() (done func()) {
	t := time.Now()

	return func() {
		// use our h.Observe here to ensure h.h gets set in the same code
		// path as the h.h.Observe() call, otherwise possibly racey.
		h.ObserveDuration(time.Since(t))
	}
}
