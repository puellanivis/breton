package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	//pb "github.com/prometheus/client_model/go"
)

type CounterValue interface {
	prometheus.Counter
}

func Counter(name string, help string, options ...Option) CounterValue {
	if !validName.MatchString(name) {
		panic("invalid metric name")
	}

	cfg := new(Options)
	for _, opt := range options {
		// in initialization, we throw the reversing Option away
		_ = opt(cfg)
	}

	// if no labels were set, then cfg.labels == nil, which
	// is how prometheus client is expecting things.
	c := prometheus.NewCounter(prometheus.CounterOpts{
		Name:        name,
		Help:        help,
		ConstLabels: cfg.constLabels,
	})

	prometheus.MustRegister(c)

	return c
}
