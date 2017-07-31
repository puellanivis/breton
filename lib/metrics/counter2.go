package metrics

import (
	"github.com/puellanivis/breton/lib/metrics/internal/atomic"
	//"github.com/prometheus/client_golang/prometheus"
)

// A ICounterValue holds the tracking information for a specific Counter or a “Child” of a Counter.
type ICounterValue struct {
	// we want to duplicate this every WithLabels() call,
	// so we don’t use a pointer here.
	metric

	c atomic.Counter
}

// WithLabels provides access to a labeled dimension of the metric, and returns a “Child” wherin the given labels are set.
// The “Child” returned is cacheable by the Caller, so as to avoid having to look it up again—this matters in latency-critical code.
func (c ICounterValue) WithLabels(labels ...Labeler) *ICounterValue {
	c.labels = c.labels.With(labels...)

	return &c
}

// Remove will remove a “Child” that matches the given labels from the metric, no longer exporting it.
func (c *ICounterValue) Remove(labels ...Labeler) bool {
	return false
}

// Clear removes all “Children” from the metric.
func (c *ICounterValue) Clear() {
	return
}

// Counter is a monotonically increasing value.
func ICounter(name string, help string, options ...Option) *ICounterValue {
	m := newMetric(name, help)

	for _, opt := range options {
		// in initialization, we throw the reverting function away
		_ = opt(m)
	}

	c := &ICounterValue{
		metric: *m,
	}

	return c
}

// Inc increments the Counter by 1.
func (c *ICounterValue) Inc() {
	c.c.Inc()
}

// Add increments the Counter by the given value.
//
// (Caller MUST NOT give a negative value.)
func (c *ICounterValue) Add(v float64) {
	c.c.Add(uint64(v))
}
