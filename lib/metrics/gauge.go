package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	pb "github.com/prometheus/client_model/go"
)

type GaugeValue struct {
	// we want to duplicate this every WithLabels() call,
	// so we don’t use a pointer here.
	metric

	g  prometheus.Gauge
	gv *prometheus.GaugeVec
}

func (g *GaugeValue) collect() <-chan prometheus.Metric {
	ch := make(chan prometheus.Metric)

	go func() {
		defer close(ch)

		if g.gv == nil {
			ch <- g.g
			return
		}

		g.gv.Collect(ch)
	}()

	return ch
}

func (g *GaugeValue) String() string {
	var list []string

	for m := range g.collect() {
		data := new(pb.Metric)
		_ = m.Write(data)

		list = append(list, fmt.Sprintf("\nmetric:<%s>", data.String()))
	}
	list = append(list, "\n")

	return fmt.Sprintf("%v", list)
}

func (g GaugeValue) WithLabels(labels ...Labeler) *GaugeValue {
	// we are working with a new copy, so no mutex is necessary.
	g.g = nil

	g.labels = g.labels.WithLabels(labels...)

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

func Gauge(name string, help string, options ...option) *GaugeValue {
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
		g.registerer.MustRegister(g.gv)

	} else {
		g.g = prometheus.NewGauge(opts)
		g.registerer.MustRegister(g.g)
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
