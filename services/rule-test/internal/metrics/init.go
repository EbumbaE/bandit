package metrics

import (
	"context"
	"net/http"

	"github.com/EbumbaE/bandit/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var server *http.Server

func InitMetricsServer(ctx context.Context) {
	mux := http.NewServeMux()
	mux.Handle("rule-test/metrics", promhttp.Handler())

	server = &http.Server{Addr: ":8080", Handler: mux}
	go func() {
		logger.Info("metrics server begin")

		go func() {
			if err := server.ListenAndServe(); err != nil {
				logger.Error("metrics server listen and serve: ", zap.Error(err))
			}
		}()
	}()
}

func StopeMetricsServer(ctx context.Context) {
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("metrics server shutdown: ", zap.Error(err))
	} else {
		logger.Info("metrics server end")
	}
}
