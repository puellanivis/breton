// Package metrics provides an abstracted metrics library compatible with prometheus client specifications.
//
// This package works very much like the standard flag library:
//	var (
//		counter   = metrics.Counter("counter_name", "usage information")
//		gauge     = metrics.Gauge("gauge_name", "usage information")
//		histogram = metrics.Histogram("histogram_name", "usage informationn") // default buckets
//		summary   = metrics.Summary("summary_name", "usage information")      // no objectives
//	)
//
// Setting up timing for a function is hopefuly straight-forward.
//	func httpHandler(w http.ResponseWriter, r *http.Request) {
//		done := summary.Timer()
//		defer done()
//
//		// do work here.
//	}
//
// A set of common 50-90-95 Summary objectives is available:
//	var summary = metrics.Summary("summary", "usage", metrics.CommonObjectives())
//
// Defining labels for a metric is hopefully also straight-forward:
//	const (
//		labelCode = metrics.Label("code")
//	)
//
//	var counter = metrics.Coutner("http_status", "usage", metrics.WithLabels(labelCode))
//
//	func httpError(w http.ResponseWriter, error string, code int) {
//		label := labelCode.WithValue(strconv.Itoa(code))
//		counter.WithLabels(label).Inc()
//
//		http.Error(w, error, code)
//	}
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
