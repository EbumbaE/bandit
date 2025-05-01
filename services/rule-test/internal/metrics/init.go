package metrics

import (
	"context"
	"net/http"

	"github.com/EbumbaE/bandit/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var server *http.Server

func StartMetricsServer(ctx context.Context, host string) {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server = &http.Server{
		Addr:    host,
		Handler: mux,
	}

	go func() {
		logger.Info("Starting metrics server", zap.String("address", server.Addr))

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
