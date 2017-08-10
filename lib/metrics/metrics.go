// Package metrics provides an abstracted metrics library compatible with prometheus client specifications.
package metrics

import (
	"fmt"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	//pb "github.com/prometheus/client_model/go"
)

// Observer is implemented by any value that has an Observe(float64) format.
// The primary metric types implementing this are Summary and Histogram.
type Observer interface {
	Observe(float64)
}

// Timer is implemented by any value that permits timing a piece of code.
type Timer interface {
	Timer() (done func())
}

type metric struct {
	registry *prometheus.Registry

	name, help string

	labels *labelScope

	*summarySettings
	*histogramSettings
}

type summarySettings struct {
	objectives map[float64]float64
	maxAge     time.Duration
	ageBuckets uint32
	bufCap     uint32
}

type histogramSettings struct {
	buckets []float64
}

var validName = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)

func newMetric(name, help string) *metric {
	if !validName.MatchString(name) {
		panic("invalid metric name")
	}

	return &metric{
		registry: prometheus.DefaultRegisterer.(*prometheus.Registry),
		name:     name,
		help:     help,
	}
}

func (m metric) helpString() string {
	if m.help == "" {
		return ""
	}

	return fmt.Sprintf("# HELP %s %s\n", m.name, m.help)
}
