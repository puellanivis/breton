package metricshttp

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	http.Handle("/metrics/prometheus", promhttp.Handler())
}
