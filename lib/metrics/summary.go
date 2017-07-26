package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// A SummaryValue holds the tracking information for a specific Summary or a “Child” of a Summary.
type SummaryValue struct {
	// we want to duplicate this every WithLabels() call,
	// so we don’t use a pointer here.
	metric

	s  prometheus.Summary
	sv *prometheus.SummaryVec
}

// WithLabels provides access to a labeled dimension of the metric, and returns a “Child” wherein the given labels are set.
// The “Child” returned is cacheable by the Caller, so as to avoid having to look it up again—this matters in latency-critical code.
func (s SummaryValue) WithLabels(labels ...Labeler) *SummaryValue {
	// we are working with a new copy, so no mutex is necessary.
	s.s = nil

	s.labels = s.labels.With(labels...)

	return &s
}

// Remove will remove a “Child” that matches the given labels from the metric, no longer exporting it.
func (s *SummaryValue) Remove(labels ...Labeler) bool {
	if s.sv == nil {
		return false
	}

	return s.sv.Delete(s.labels.getMap())
}

// Clear removes all “Children” from the metric.
func (s *SummaryValue) Clear() {
	if s.sv != nil {
		s.sv.Reset()
	}
}

// Summary samples observations (usually things like request durations) over sliding windows of time,
// and provides instantaneous insight into their distributions, frequencies, and sums.
func Summary(name string, help string, options ...Option) *SummaryValue {
	m := newMetric(name, help)

	// prometheus library has deprecated DefaultObjectives, as the
	// implementation documentation says that Summaries MUST allow for not
	// having quantiles, and that this MUST be the default.
	//
	// As such, we set a default empty map here to override any default
	// currently in use by the wrapped prometheus library.
	m.summarySettings = &summarySettings{
		objectives: make(map[float64]float64),
	}

	for _, opt := range options {
		// in initialization, we throw the reverting function away
		_ = opt(m)
	}

	s := &SummaryValue{
		metric: *m,
	}

	opts := prometheus.SummaryOpts{
		Name:       name,
		Help:       help,
		Objectives: m.objectives,
		MaxAge:     m.maxAge,
		AgeBuckets: m.ageBuckets,
		BufCap:     m.bufCap,
	}

	if s.labels != nil {
		if _, ok := m.labels.set.canSet["quantile"]; ok {
			panic("summaries cannot allow \"quantile\" as a label")
		}

		s.sv = prometheus.NewSummaryVec(opts, s.labels.set.keys)
		s.registry.MustRegister(s.sv)

	} else {
		s.s = prometheus.NewSummary(opts)
		s.registry.MustRegister(s.s)
	}

	return s
}

// Observe records the given value into the Summary.
func (s *SummaryValue) Observe(v float64) {
	if s.s == nil {
		// function is idempotent, and won’t step on others’ toes
		s.s = s.sv.With(s.labels.getMap()).(prometheus.Summary)
	}

	s.s.Observe(v)
}

// ObserveDuration records the given Duration into the Summary.
func (s *SummaryValue) ObserveDuration(d time.Duration) {
	s.Observe(d.Seconds())
}

// Timer times a piece of code and records to the Summary its duration in seconds.
//
// (Caller MUST ensure the returned done function is called, and SHOULD use defer.)
func (s *SummaryValue) Timer() (done func()) {
	t := time.Now()

	return func() {
		// use our s.Observe here to ensure s.s gets set in the same code
		// path as the s.s.Observe() call, otherwise possibly racey.
		s.ObserveDuration(time.Since(t))
	}
}
