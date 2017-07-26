package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// An Option defines specific optional feature for a Metric. When applied,
// it returns a new Option that will revert the feature to the previous state.
type Option func(m *metric) Option

// withLabelScope allows for toggling labelScopes directly, and not just a list of Labelers to keep enscoping into.
func withLabelScope(labels *labelScope) Option {
	return func(m *metric) Option {
		save := m.labels

		m.labels = labels

		return withLabelScope(save)
	}
}

// WithLabels defines a set of Labels to use for a Metric.
// * Label: a Label with the given name, no default.
// * Label.WithValue: a Label with the given name, and default value.
// * Label.Const; a Label with the given name, and a constant value.
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

// WithRegistry defines the Registry to which a Metric will be registered.
// By default, every Metric will be registered with the default prometheus.Registery.
func WithRegistry(registry *prometheus.Registry) Option {
	return func(m *metric) Option {
		save := m.registry

		m.registry = registry

		return WithRegistry(save)
	}
}

// LinearBuckets defines a series of linear buckets defined by:
//	for i from 0 to count: a_i = start + width × i
// (Caller MUST NOT pass a count <= 0)
func LinearBuckets(start, width float64, count uint) Option {
	return WithBuckets(prometheus.LinearBuckets(start, width, int(count))...)
}

// ExponentialBuckets defines a series of exponential buckets defined by:
// 	for i from 0 to count: a_i = start × factor^i
// (Caller MUST NOT pass a count <= 0, start <= 0, or factor <= 1)
func ExponentialBuckets(start, factor float64, count uint) Option {
	return WithBuckets(prometheus.ExponentialBuckets(start, factor, int(count))...)
}

// WithBuckets defines the buckets into which observations are counted. Each
// element in the slice is the upper inclusive bound of a bucket.
// (Caller MUST ensure that the buckets are defined in increasing order.)
// (Caller MUST not include the highest +Inf bucket boundary, it is added
// implicity.)
func WithBuckets(buckets ...float64) Option {
	return func(m *metric) Option {
		if m.histogramSettings == nil {
			panic("metric is not a histogram")
		}

		save := m.buckets

		m.buckets = buckets

		return WithBuckets(save...)
	}

}

// WithObjectives defines the quantile rank estimates with their respective
// absolute error. If Objectives[q] = e, then the value reported for q will be
// the φ-quantile value for some φ in the range q±e.
// The default is to have no Objectives.
// For a common case of 50-90-99th percentiles, use CommonObjectives
func WithObjectives(objectives map[float64]float64) Option {
	return func(m *metric) Option {
		if m.summarySettings == nil {
			panic("metric is not a summary")
		}

		save := m.objectives

		m.objectives = objectives

		return WithObjectives(save)
	}
}

var (
	commonObjectives = map[float64]float64{
		0.5:  0.05,
		0.9:  0.01,
		0.99: 0.001,
	}
)

// CommonObjectives defines a common set of Objectives, which track the
// percentiles: 50±5 (the median), 90±1, and 99±0.1
// By default, a Summary will not have any Objectives.
func CommonObjectives() Option {
	return WithObjectives(commonObjectives)
}

// WithMaxAge defines the duration for which an observation stays relevant
// for the summary. Must be positive.
func WithMaxAge(value time.Duration) Option {
	if value <= 0 {
		panic("value for maximum age must be positive")
	}

	return func(m *metric) Option {
		if m.summarySettings == nil {
			panic("metric is not a summary")
		}

		save := m.maxAge

		m.maxAge = value

		return WithMaxAge(save)
	}
}

// WithAgeBuckets is the number of buckets used to exclude observations that
// are older than MaxAge from the summary. A higher number has a resource
// penalty, so only increase it if the higher resolution is really required.
// For very high observation rates, you might want to reduce the number of
// age buckets. With only one age bucket, you will effectively see a complete
// reset of the summary each time MaxAge has passed.
func WithAgeBuckets(value uint32) Option {
	return func(m *metric) Option {
		if m.summarySettings == nil {
			panic("metric is not a summary")
		}

		save := m.ageBuckets

		m.ageBuckets = value

		return WithAgeBuckets(save)
	}
}

// WithBufCap deifnes the default sample stream buffer size. The default value
// should suffice for most users. If there is a need to increase the value,
// a multiple of 500 is recommended (because that is the internal buffer size
// of the underlying package "github.com/bmizerany/perks/quantile")
//
// This Option exposes implementation details, which is undesirable.
func WithBufCap(value uint32) Option {
	return func(m *metric) Option {
		if m.summarySettings == nil {
			panic("metric is not a summary")
		}

		save := m.bufCap

		m.bufCap = value

		return WithBufCap(save)
	}
}
