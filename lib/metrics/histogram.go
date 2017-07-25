package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	pb "github.com/prometheus/client_model/go"
)

type HistogramValue struct {
	// we want to duplicate this every WithLabels() call,
	// so we don’t use a pointer here.
	metric

	h  prometheus.Histogram
	hv *prometheus.HistogramVec
}

func (h *HistogramValue) collect() <-chan prometheus.Metric {
	ch := make(chan prometheus.Metric)

	go func() {
		defer close(ch)

		if h.hv == nil {
			ch <- h.h
			return
		}

		h.hv.Collect(ch)
	}()

	return ch
}

func (h *HistogramValue) String() string {
	var list []string

	for m := range h.collect() {
		data := new(pb.Metric)
		_ = m.Write(data)

		list = append(list, fmt.Sprintf("\nmetric:<%s>", data.String()))
	}
	list = append(list, "\n")

	return fmt.Sprintf("%v", list)
}

func (h HistogramValue) WithLabels(labels ...Labeler) *HistogramValue {
	// we are working with a new copy, so no mutex is necessary.
	h.h = nil

	h.labels = h.labels.WithLabels(labels...)

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

func Histogram(name string, help string, options ...option) *HistogramValue {
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
		h.registerer.MustRegister(h.hv)

	} else {
		h.h = prometheus.NewHistogram(opts)
		h.registerer.MustRegister(h.h)
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
