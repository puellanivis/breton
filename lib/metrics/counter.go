package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type CounterValue struct {
	// we want to duplicate this every WithLabels() call,
	// so we don’t use a pointer here.
	metric

	c  prometheus.Counter
	cv *prometheus.CounterVec
}

func (c CounterValue) WithLabels(labels ...Labeler) *CounterValue {
	// we are working with a new copy, so no mutex is necessary.
	c.c = nil

	c.labels = c.labels.With(labels...)

	return &c
}

func (c *CounterValue) Reset() {
	if c.cv != nil {
		c.cv.Reset()
	}
}

func (c *CounterValue) Delete(labels ...Labeler) bool {
	if c.cv == nil {
		return false
	}

	return c.cv.Delete(c.labels.getMap())
}

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

func (c *CounterValue) Inc() {
	if c.c == nil {
		// function is idempotent, and won’t step on others’ toes
		c.c = c.cv.With(c.labels.getMap())
	}

	c.c.Inc()
}

func (c *CounterValue) Add(v float64) {
	if c.c == nil {
		// function is idempotent, and won’t step on others’ toes
		c.c = c.cv.With(c.labels.getMap())
	}

	c.c.Add(v)
}
