package metrics

import (
	//"sort"
	//"sync"
	//"time"

	"github.com/prometheus/client_golang/prometheus"
)

// An Option is a function that applies a specific option to a Metric, and returns an Option that will revert the change.
type Option func(m *metric) Option

// withLabelScope allows for toggling labelScopes directly, and not just a list of Labelers to keep enscoping into.
func withLabelScope(labels *labelScope) Option {
	return func(m *metric) Option {
		save := m.labels

		m.labels = labels

		return withLabelScope(save)
	}
}

// WithLabels returns an Option that will add a set of Labels to a Metric.
func WithLabels(labels ...Labeler) Option {
	return func(m *metric) Option {
		save := m.labels

		if save == nil {
			m.labels = defineLabels(labels...)

		} else {
			m.labels = m.labels.With(labels...)
		}

		return withLabelScope(save)
	}
}

// WithRegistry returns an Option that will switch the Registry to which a Metric will be registered.
func WithRegistry(registry *prometheus.Registry) Option {
	return func(m *metric) Option {
		save := m.registry

		m.registry = registry

		return WithRegistry(save)
	}
}

// LinearBuckets returns an Option that allows one to define a series of linear buckets defined by:
//	for i from 0 to count: a_i = start + width × i
func LinearBuckets(start, width float64, count uint) Option {
	return WithBuckets(prometheus.LinearBuckets(start, width, int(count))...)
}

// ExponentialBuckets returns an Option that allows one to define a series of exponential buckets defined by:
// 	for i from 0 to count: a_i = start × factor^i
func ExponentialBuckets(start, factor float64, count uint) Option {
	return WithBuckets(prometheus.ExponentialBuckets(start, factor, int(count))...)
}

// WithBuckets returns an Option that allows one to define a series of arbitrary bucket values.
// (Caller MUST ensure that the buckets are defined in increasing order.)
func WithBuckets(buckets ...float64) Option {
	return func(m *metric) Option {
		save := m.buckets

		m.buckets = buckets

		return WithBuckets(save...)
	}

}

// WithObjectives returns an Option that allows one to define a set of Objectives for a Summary.
//
// Reference: https://prometheus.io/docs/concepts/metric_types/#summary
func WithObjectives(objectives map[float64]float64) Option {
	return func(m *metric) Option {
		save := m.objectives

		m.objectives = objectives

		return WithObjectives(save)
	}
}
