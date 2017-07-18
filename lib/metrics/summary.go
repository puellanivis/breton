package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	//pb "github.com/prometheus/client_model/go"
)

type SummaryValue struct {
	prometheus.Summary
}

func (s *SummaryValue) Timer() *Timer {
	return newTimer(s)
}

func Summary(name string, help string, opts ...Option) *SummaryValue {
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
	c := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:        name,
		Help:        help,
		ConstLabels: cfg.constLabels,
		Objectives:  cfg.objectives,
		MaxAge:      cfg.maxAge,
		AgeBuckets:  uint32(cfg.ageBuckets),
		BufCap:      uint32(cfg.bufCap),
	})

	prometheus.MustRegister(c)

	return &SummaryValue{c}
}
