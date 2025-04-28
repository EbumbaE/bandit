package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	DataCounter         *prometheus.CounterVec
	SummaryResponceTime prometheus.Summary
)

func init() {
	DataCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "rule_test",
			Name:      "data_count",
		},
		[]string{"data"},
	)
	SummaryResponceTime = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "rule_test",
		Name:      "data_responce_time_ms",
		Objectives: map[float64]float64{
			0.5:  10,
			0.9:  20,
			0.99: 50,
		},
	})
}
