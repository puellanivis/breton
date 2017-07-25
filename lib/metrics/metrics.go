package metrics

import (
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	//pb "github.com/prometheus/client_model/go"
)

type Observer interface {
	Observe(float64)
}

type TimeKeeper interface {
	Timer() *Timer
}

type metric struct {
	registerer prometheus.Registerer

	name, help string

	labels *Labels

	objectives map[float64]float64
	buckets []float64
}

var validName = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)

func newMetric(name, help string) *metric {
	if !validName.MatchString(name) {
		panic("invalid metric name")
	}

	return &metric{
		registerer: prometheus.DefaultRegisterer,
		name:       name,
		help:       help,
	}
}

func (m metric) Name() string {
	return m.name
}

func (m metric) Help() string {
	if m.help == "" {
		return ""
	}

	return fmt.Sprintf("# HELP %s %s\n", m.name, m.help)
}
