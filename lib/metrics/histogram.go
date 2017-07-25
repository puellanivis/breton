package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type HistogramValue struct {
	// we want to duplicate this every WithLabels() call,
	// so we don’t use a pointer here.
	metric

	h  prometheus.Histogram
	hv *prometheus.HistogramVec
}

func (h HistogramValue) WithLabels(labels ...Labeler) *HistogramValue {
	// we are working with a new copy, so no mutex is necessary.
	h.h = nil

	h.labels = h.labels.With(labels...)

	return &h
}

func (h *HistogramValue) Reset() {
	if h.hv != nil {
		h.hv.Reset()
	}
}

func (h *HistogramValue) Delete(labels ...Labeler) bool {
	if h.hv == nil {
		return false
	}

	return h.hv.Delete(h.labels.getMap())
}

func Histogram(name string, help string, options ...Option) *HistogramValue {
	m := newMetric(name, help)

	for _, opt := range options {
		// in initialization, we throw the reverting function away
		_ = opt(m)
	}

	h := &HistogramValue{
		metric: *m,
	}

	opts := prometheus.HistogramOpts{
		Name: name,
		Help: help,
		Buckets: m.buckets,
	}

	if h.labels != nil {
		h.hv = prometheus.NewHistogramVec(opts, h.labels.set.keys)
		h.registry.MustRegister(h.hv)

	} else {
		h.h = prometheus.NewHistogram(opts)
		h.registry.MustRegister(h.h)
	}

	return h
}

func (h *HistogramValue) Observe(v float64) {
	if h.h == nil {
		// function is idempotent, and won’t step on others’ toes
		h.h = h.hv.With(h.labels.getMap()).(prometheus.Histogram)
	}

	h.h.Observe(v)
}

func (h *HistogramValue) Timer() *Timer {
	if h.h == nil {
		// function is idempotent, and won’t step on others’ toes
		h.h = h.hv.With(h.labels.getMap()).(prometheus.Histogram)
	}

	return newTimer(h.h)
}
