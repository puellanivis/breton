package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// A CounterValue holds the tracking information for a specific Counter or a “Child” of a Counter.
type CounterValue struct {
	// we want to duplicate this every WithLabels() call,
	// so we don’t use a pointer here.
	metric

	c  prometheus.Counter
	cv *prometheus.CounterVec
}

// WithLabels provides access to a labeled dimension of the metric, and returns a “Child” wherin the given labels are set.
// The “Child” returned is cacheable by the Caller, so as to avoid having to look it up again—this matters in latency-critical code.
func (c CounterValue) WithLabels(labels ...Labeler) *CounterValue {
	// we are working with a new copy, so no mutex is necessary.
	c.c = nil

	c.labels = c.labels.With(labels...)

	return &c
}

// Remove will remove a “Child” that matches the given labels from the metric, no longer exporting it.
func (c *CounterValue) Remove(labels ...Labeler) bool {
	if c.cv == nil {
		return false
	}

	return c.cv.Delete(c.labels.getMap())
}

// Clear removes all “Children” from the metric.
func (c *CounterValue) Clear() {
	if c.cv != nil {
		c.cv.Reset()
	}
}

// Counter is a monotonically increasing value.
func Counter(name string, help string, options ...Option) *CounterValue {
	m := newMetric(name, help)

	for _, opt := range options {
		// in initialization, we throw the reverting function away
		_ = opt(m)
	}

	c := &CounterValue{
		metric: *m,
	}

	opts := prometheus.CounterOpts{
		Name: name,
		Help: help,
	}

	if c.labels != nil {
		c.cv = prometheus.NewCounterVec(opts, c.labels.set.keys)
		c.registry.MustRegister(c.cv)

	} else {
		c.c = prometheus.NewCounter(opts)
		c.registry.MustRegister(c.c)
	}

	return c
}

// Inc increments the Counter by 1.
func (c *CounterValue) Inc() {
	if c.c == nil {
		// function is idempotent, and won’t step on others’ toes
		c.c = c.cv.With(c.labels.getMap())
	}

	c.c.Inc()
}

// Add increments the Counter by the given value.
// The given value MUST NOT be negative.
func (c *CounterValue) Add(v float64) {
	if v < 0 {
		panic("counter cannot decrease in value")
	}

	if c.c == nil {
		// function is idempotent, and won’t step on others’ toes
		c.c = c.cv.With(c.labels.getMap())
	}

	c.c.Add(v)
}
