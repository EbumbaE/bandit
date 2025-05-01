package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	DataCounter  *prometheus.CounterVec
	ResponceTime *prometheus.HistogramVec
)

func init() {
	DataCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "rule_test",
			Name:      "data_count",
		},
		[]string{"data"},
	)
	ResponceTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_handler_duration_ms",
			Help:    "Duration of gRPC handler execution in ms",
			Buckets: []float64{0.5, 1, 2, 3, 5, 10, 12, 15, 20, 50, 100},
		},
		[]string{"handler", "status"},
	)
}
