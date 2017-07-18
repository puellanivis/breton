package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	//pb "github.com/prometheus/client_model/go"
)

type HistogramValue struct {
	prometheus.Histogram
}

func (h *HistogramValue) Timer() *Timer {
	return newTimer(h)
}

func Histogram(name string, help string, opts ...Option) *HistogramValue {
	if !validName.MatchString(name) {
		panic("invalid metric name")
	}

	cfg := new(Options)
	for _, opt := range opts {
		// in initialization, we throw the reversing Option away
		_ = opt(cfg)
	}

	// if no labels were set, then cfg.labels == nil, which
	// is how prometheus client is expecting things.
	c := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:        name,
		Help:        help,
		ConstLabels: cfg.constLabels,
		Buckets:     cfg.buckets,
	})

	prometheus.MustRegister(c)

	return &HistogramValue{c}
}
