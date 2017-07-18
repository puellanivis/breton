package metrics

import (
	"sort"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Options struct {
	makeConstLabels sync.Once
	constLabels     prometheus.Labels
	labels          []string

	// histograms only
	buckets []float64

	// summaries only
	makeObjectives sync.Once
	objectives     map[float64]float64

	maxAge time.Duration
	// prometheus client expects uint32 for these, but Go guarantees unsafe.SizeOf(uint) >= unsafe.SizeOf(uint32)
	// implementation-details-wise, we can cast this when necessary... i.e. at creation time
	ageBuckets, bufCap uint
}

type Option func(opt *Options) Option

func noopOption() Option {
	return func(opt *Options) Option {
		return noopOption()
	}
}

func withoutConstLabel(name string) Option {
	return func(opt *Options) Option {
		if opt == nil {
			return noopOption()
		}

		save, ok := opt.constLabels[name]
		if !ok {
			return noopOption()
		}

		delete(opt.constLabels, name)
		return WithConstLabel(name, save)
	}
}

// WithLabel returns an Option function that will add a label, value pair to a Metric, and return an Option that will undo that change.
func WithConstLabel(name, value string) Option {
	return func(opt *Options) Option {
		opt.makeConstLabels.Do(func() {
			opt.constLabels = make(prometheus.Labels)
		})

		save, ok := opt.constLabels[name]
		opt.constLabels[name] = value

		if !ok {
			return withoutConstLabel(name)
		}

		return WithConstLabel(name, save)
	}
}

func withoutObjective(quantile float64) Option {
	return func(opt *Options) Option {
		if opt == nil {
			return noopOption()
		}

		save, ok := opt.objectives[quantile]
		if !ok {
			return noopOption()
		}

		delete(opt.objectives, quantile)
		return WithObjective(quantile, save)
	}
}

// WithObjective returns an Option function that will add a objective, value pair to a Metric, and return an Option that will undo that change.
func WithObjective(quantile, epsilon float64) Option {
	return func(opt *Options) Option {
		opt.makeObjectives.Do(func() {
			opt.objectives = make(map[float64]float64)
		})

		save, ok := opt.objectives[quantile]
		opt.objectives[quantile] = epsilon

		if !ok {
			return withoutObjective(quantile)
		}

		return WithObjective(quantile, save)
	}
}

// WithLabels returns an Option function that will add a label, value pair to a Metric, and return an Option that will undo that change.
func WithLabels(labels ...string) Option {
	return func(opt *Options) Option {
		save := opt.labels
		opt.labels = labels

		return WithLabels(save...)
	}
}

// WithBuckets returns an Option function that will add a label, value pair to a Metric, and return an Option that will undo that change.
func WithBuckets(buckets ...float64) Option {
	if !sort.Float64sAreSorted(buckets) {
		panic("buckets must be in strictly ascending order")
	}

	return func(opt *Options) Option {
		save := opt.buckets
		opt.buckets = buckets

		return WithBuckets(save...)
	}
}

// WithBucketsLinear returns an Option function that will add a label, value pair to a Metric, and return an Option that will undo that change.
// This function is so named, so that it appears in documentation with the WithBuckets...() functions
func WithBucketsLinear(start, width float64, count uint) Option {
	var buckets []float64

	for i := uint(0); i < count; i++ {
		buckets = append(buckets, start)
		start += width
	}

	return WithBuckets(buckets...)
}

// WithBucketsExponential returns an Option function that will add a label, value pair to a Metric, and return an Option that will undo that change.
// This function is so named, so that it appears in documentation with the WithBuckets...() functions
func WithBucketsExponential(start, factor float64, count uint) Option {
	var buckets []float64

	if start <= 0 {
		panic("need a positive start value")
	}

	if factor <= 1 {
		panic("need a factor greater than 1")
	}

	for i := uint(0); i < count; i++ {
		buckets = append(buckets, start)
		start *= factor
	}

	return WithBuckets(buckets...)
}

func WithMaxAge(value time.Duration) Option {
	return func(opt *Options) Option {
		save := opt.maxAge
		opt.maxAge = value

		return WithMaxAge(save)
	}
}

func WithAgeBuckets(value uint) Option {
	return func(opt *Options) Option {
		save := opt.ageBuckets
		opt.ageBuckets = value

		return WithAgeBuckets(save)
	}
}

func WithBufferCap(value uint) Option {
	return func(opt *Options) Option {
		save := opt.bufCap
		opt.bufCap = value

		return WithBufferCap(save)
	}
}
