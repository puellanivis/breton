package metrics

import (
	//"sort"
	//"sync"
	//"time"

	"github.com/prometheus/client_golang/prometheus"
)

type option func(m *metric) option

func withLabelsObj(labels *Labels) option {
	return func(m *metric) option {
		save := m.labels

		m.labels = labels

		return withLabelsObj(save)
	}
}

// WithLabel returns an option function that will add a label, value pair to a Metric, and return an option that will undo that change.
func WithLabels(labels ...Labeler) option {
	return func(m *metric) option {
		save := m.labels

		if save == nil {
			m.labels = DefineLabels(labels...)

		} else {
			m.labels = m.labels.WithLabels(labels...)
		}

		return withLabelsObj(save)
	}
}

// WithRegistry returns an option function that will switch the registry to which a metric will be registered.
func WithRegistry(registerer prometheus.Registerer) option {
	return func(m *metric) option {
		save := m.registerer

		m.registerer = registerer

		return WithRegistry(save)
	}
}

func WithLinear(start, width float64, count int) option {
	return WithBuckets(prometheus.LinearBuckets(start, width, count)...)
}

func WithExponential(start, factor float64, count int) option {
	return WithBuckets(prometheus.ExponentialBuckets(start, factor, count)...)
}

func WithBuckets(buckets ...float64) option {
	return func(m *metric) option {
		save := m.buckets

		m.buckets = buckets

		return WithBuckets(save...)
	}

}

func WithObjectives(objectives map[float64]float64) option {
	return func(m *metric) option {
		save := m.objectives

		m.objectives = objectives

		return WithObjectives(save)
	}
}
