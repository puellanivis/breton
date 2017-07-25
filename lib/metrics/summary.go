package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type SummaryValue struct {
	// we want to duplicate this every WithLabels() call,
	// so we don’t use a pointer here.
	metric

	s  prometheus.Summary
	sv *prometheus.SummaryVec
}

func (s SummaryValue) WithLabels(labels ...Labeler) *SummaryValue {
	// we are working with a new copy, so no mutex is necessary.
	s.s = nil

	s.labels = s.labels.With(labels...)

	return &s
}

func (s *SummaryValue) Reset() {
	if s.sv != nil {
		s.sv.Reset()
	}
}

func (s *SummaryValue) Delete(labels ...Labeler) bool {
	if s.sv == nil {
		return false
	}

	return s.sv.Delete(s.labels.getMap())
}

func Summary(name string, help string, options ...Option) *SummaryValue {
	m := newMetric(name, help)

	for _, opt := range options {
		// in initialization, we throw the reverting function away
		_ = opt(m)
	}

	s := &SummaryValue{
		metric: *m,
	}

	opts := prometheus.SummaryOpts{
		Name: name,
		Help: help,
		Objectives: m.objectives,
		// TODO: MaxAge
		// TODO: AgeBuckets
		// TODO: BufCap
	}

	if s.labels != nil {
		s.sv = prometheus.NewSummaryVec(opts, s.labels.set.keys)
		s.registry.MustRegister(s.sv)

	} else {
		s.s = prometheus.NewSummary(opts)
		s.registry.MustRegister(s.s)
	}

	return s
}

func (s *SummaryValue) Observe(v float64) {
	if s.s == nil {
		// function is idempotent, and won’t step on others’ toes
		s.s = s.sv.With(s.labels.getMap()).(prometheus.Summary)
	}

	s.s.Observe(v)
}

func (s *SummaryValue) Timer() *Timer {
	if s.s == nil {
		// function is idempotent, and won’t step on others’ toes
		s.s = s.sv.With(s.labels.getMap()).(prometheus.Summary)
	}

	return newTimer(s.s)
}
