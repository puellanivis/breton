package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	//pb "github.com/prometheus/client_model/go"
)

type Observer interface {
	Observe(float64)
}

type TimeKeeper interface {
	Timer() *Timer
}

type Metric prometheus.Metric
