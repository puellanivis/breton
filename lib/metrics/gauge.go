package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	//pb "github.com/prometheus/client_model/go"
)

type GaugeValue prometheus.Gauge

func Gauge(name string, help string, opts ...Option) GaugeValue {
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
	c := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        name,
		Help:        help,
		ConstLabels: cfg.constLabels,
	})

	prometheus.MustRegister(c)

	return c
}
