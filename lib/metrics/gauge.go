package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type GaugeValue struct {
	// we want to duplicate this every WithLabels() call,
	// so we don’t use a pointer here.
	metric

	g  prometheus.Gauge
	gv *prometheus.GaugeVec
}

func (g GaugeValue) WithLabels(labels ...Labeler) *GaugeValue {
	// we are working with a new copy, so no mutex is necessary.
	g.g = nil

	g.labels = g.labels.With(labels...)

	return &g
}

func (g *GaugeValue) Reset() {
	if g.gv != nil {
		g.gv.Reset()
	}
}

func (g *GaugeValue) Delete(labels ...Labeler) bool {
	if g.gv == nil {
		return false
	}

	return g.gv.Delete(g.labels.getMap())
}

func Gauge(name string, help string, options ...Option) *GaugeValue {
	m := newMetric(name, help)

	for _, opt := range options {
		// in initialization, we throw the reverting function away
		_ = opt(m)
	}

	g := &GaugeValue{
		metric: *m,
	}

	opts := prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}

	if g.labels != nil {
		g.gv = prometheus.NewGaugeVec(opts, g.labels.set.keys)
		g.registry.MustRegister(g.gv)

	} else {
		g.g = prometheus.NewGauge(opts)
		g.registry.MustRegister(g.g)
	}

	return g
}

func (g *GaugeValue) Inc() {
	if g.g == nil {
		// function is idempotent, and won’t step on others’ toes
		g.g = g.gv.With(g.labels.getMap())
	}

	g.g.Inc()
}

func (g *GaugeValue) Dec() {
	if g.g == nil {
		// function is idempotent, and won’t step on others’ toes
		g.g = g.gv.With(g.labels.getMap())
	}

	g.g.Dec()
}

func (g *GaugeValue) Add(v float64) {
	if g.g == nil {
		// function is idempotent, and won’t step on others’ toes
		g.g = g.gv.With(g.labels.getMap())
	}

	g.g.Add(v)
}

func (g *GaugeValue) Sub(v float64) {
	if g.g == nil {
		// function is idempotent, and won’t step on others’ toes
		g.g = g.gv.With(g.labels.getMap())
	}

	g.g.Sub(v)
}

func (g *GaugeValue) Set(v float64) {
	if g.g == nil {
		// function is idempotent, and won’t step on others’ toes
		g.g = g.gv.With(g.labels.getMap())
	}

	g.g.Set(v)
}

func (g *GaugeValue) SetToCurrentTime() {
	if g.g == nil {
		// function is idempotent, and won’t step on others’ toes
		g.g = g.gv.With(g.labels.getMap())
	}

	g.g.SetToCurrentTime()
}
