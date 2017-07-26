package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// A GaugeValue holds the tracking information for a specific Gauge or a “Child” of a Gauge.
type GaugeValue struct {
	// we want to duplicate this every WithLabels() call,
	// so we don’t use a pointer here.
	metric

	g  prometheus.Gauge
	gv *prometheus.GaugeVec
}

// WithLabels provides access to a labeled dimension of the metric, and returns a “Child” wherein the given labels are set.
// The “Child” returned is cacheable by the Caller, so as to avoid having to look it up again—this matters in latency-critical code.
func (g GaugeValue) WithLabels(labels ...Labeler) *GaugeValue {
	// we are working with a new copy, so no mutex is necessary.
	g.g = nil

	g.labels = g.labels.With(labels...)

	return &g
}

// Remove will remove a “Child” that matches the given labels from the metric, no longer exporting it.
func (g *GaugeValue) Remove(labels ...Labeler) bool {
	if g.gv == nil {
		return false
	}

	return g.gv.Delete(g.labels.getMap())
}

// Clear removes all “Children” from the metric.
func (g *GaugeValue) Clear() {
	if g.gv != nil {
		g.gv.Reset()
	}
}

// Gauge represents a value that can go up and down.
func Gauge(name string, help string, options ...Option) *GaugeValue {
	m := newMetric(name, help)

	for _, opt := range options {
		// in initialization, we throw the reverting function away
		_ = opt(m)
	}

	g := &GaugeValue{
		metric: *m,
	}

	opts := prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}

	if g.labels != nil {
		g.gv = prometheus.NewGaugeVec(opts, g.labels.set.keys)
		g.registry.MustRegister(g.gv)

	} else {
		g.g = prometheus.NewGauge(opts)
		g.registry.MustRegister(g.g)
	}

	return g
}

// Inc increments the Gauge by 1.
func (g *GaugeValue) Inc() {
	if g.g == nil {
		// function is idempotent, and won’t step on others’ toes
		g.g = g.gv.With(g.labels.getMap())
	}

	g.g.Inc()
}

// Add increments the Gauge by the given value.
func (g *GaugeValue) Add(v float64) {
	if g.g == nil {
		// function is idempotent, and won’t step on others’ toes
		g.g = g.gv.With(g.labels.getMap())
	}

	g.g.Add(v)
}

// Dec decrements the Gauge by 1.
func (g *GaugeValue) Dec() {
	if g.g == nil {
		// function is idempotent, and won’t step on others’ toes
		g.g = g.gv.With(g.labels.getMap())
	}

	g.g.Dec()
}

// Sub decrements the Gauge by the given value.
func (g *GaugeValue) Sub(v float64) {
	if g.g == nil {
		// function is idempotent, and won’t step on others’ toes
		g.g = g.gv.With(g.labels.getMap())
	}

	g.g.Sub(v)
}

// Set sets the Gauge to the given value.
func (g *GaugeValue) Set(v float64) {
	if g.g == nil {
		// function is idempotent, and won’t step on others’ toes
		g.g = g.gv.With(g.labels.getMap())
	}

	g.g.Set(v)
}

// SetToTime sets the Gauge to the given Time in seconds.
func (g *GaugeValue) SetToTime(t time.Time) {
	g.Set(float64(t.UnixNano()) / 1e9)
}

// SetToDuration sets the Gauge to the given Duration in seconds.
func (g *GaugeValue) SetToDuration(d time.Duration) {
	g.Set(d.Seconds())
}

// Timer times a piece of code and sets the Gauge to its duration in seconds.
// This is useful for batch jobs. The Timer will commit the duration when the done function is called.
// (Caller MUST ensure the returned done function is called, and SHOULD use defer.)
func (g *GaugeValue) Timer() (done func()) {
	// get start time as fast as possible, then set the Gauge to zero.
	// reference: https://prometheus.io/docs/practices/instrumentation/#avoid-missing-metrics
	t := time.Now()
	g.Set(0)

	return func() {
		// use our g.Set here to ensure g.g gets set in the same code
		// path as the g.g.Set() call, otherwise possibly racey.
		g.SetToDuration(time.Since(t))
	}
}
