package metrics

import (
	//"sort"
	//"sync"
	//"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Option func(m *metric) Option

func withLabelScope(labels *labelScope) Option {
	return func(m *metric) Option {
		save := m.labels

		m.labels = labels

		return withLabelScope(save)
	}
}

// WithLabel returns an Option function that will add a label, value pair to a Metric, and return an Option that will undo that change.
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

// WithRegistry returns an Option function that will switch the registry to which a metric will be registered.
func WithRegistry(registry *prometheus.Registry) Option {
	return func(m *metric) Option {
		save := m.registry

		m.registry = registry

		return WithRegistry(save)
	}
}

func WithLinear(start, width float64, count int) Option {
	return WithBuckets(prometheus.LinearBuckets(start, width, count)...)
}

func WithExponential(start, factor float64, count int) Option {
	return WithBuckets(prometheus.ExponentialBuckets(start, factor, count)...)
}

func WithBuckets(buckets ...float64) Option {
	return func(m *metric) Option {
		save := m.buckets

		m.buckets = buckets

		return WithBuckets(save...)
	}

}

func WithObjectives(objectives map[float64]float64) Option {
	return func(m *metric) Option {
		save := m.objectives

		m.objectives = objectives

		return WithObjectives(save)
	}
}
