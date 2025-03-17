package server

import (
	"context"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	ruleadmin "github.com/EbumbaE/bandit/pkg/genproto/rule-admin/api"
	"github.com/EbumbaE/bandit/pkg/logger"
)

func InitRuleAdminSwagger(ctx context.Context, wg *sync.WaitGroup, swaggerAddr, swaggerHost, grpcHost string) {
	httpMux := http.NewServeMux()

	relativePath := "./pkg/genproto/rule-admin/api/rule-admin.swagger.json"
	absolutePath, err := filepath.Abs(relativePath)
	if err != nil {
		logger.Error("build absolutePath", zap.Error(err))
	}

	httpMux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, absolutePath)
	})

	httpMux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://"+swaggerHost+"/swagger.json"),
	))

	grpcMux := runtime.NewServeMux()
	if err := ruleadmin.RegisterRuleAdminServiceHandlerFromEndpoint(ctx, grpcMux, grpcHost, []grpc.DialOption{grpc.WithInsecure()}); err != nil {
		logger.Error("failed to register gateway handler", zap.Error(err))
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		srv := &http.Server{
			Addr: swaggerAddr,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/swagger") {
					httpMux.ServeHTTP(w, r)
					return
				}
				grpcMux.ServeHTTP(w, r)
			}),
		}

		logger.Info("swagger for rule-admin start")
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				logger.Error("swagger for rule-admin", zap.Error(err))
			}
		}()

		<-ctx.Done()
		_ = srv.Shutdown(ctx)

		logger.Info("swagger for rule-admin stop")
	}()
}

func StartRuleAdmin(ctx context.Context, serv ruleadmin.RuleAdminServiceServer, wg *sync.WaitGroup, host string) {
	listener, err := net.Listen("tcp", host)
	if err != nil {
		logger.Error("failed to listen in sender server", zap.Error(err))
	}
	server := grpc.NewServer()
	ruleadmin.RegisterRuleAdminServiceServer(server, serv)
	reflection.Register(server)

	wg.Add(1)
	go func() {
		logger.Info("rule-admin start")
		defer wg.Done()

		go func() {
			if err := server.Serve(listener); err != nil {
				logger.Error("failed to serve in sender server", zap.Error(err))
			}
		}()

		<-ctx.Done()
		server.GracefulStop()

		logger.Info("sender server stop")
	}()
}
