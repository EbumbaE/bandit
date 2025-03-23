package main

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/EbumbaE/bandit/pkg/logger"
	"github.com/EbumbaE/bandit/pkg/redis"

	rule_diller_service "github.com/EbumbaE/bandit/services/rule-diller/app"
	rule_diller_consumer "github.com/EbumbaE/bandit/services/rule-diller/internal/consumer"
	"github.com/EbumbaE/bandit/services/rule-diller/internal/provider"
	rule_diller_storage "github.com/EbumbaE/bandit/services/rule-diller/internal/storage"
	"github.com/EbumbaE/bandit/services/rule-diller/server"
)

type connections struct {
	redisConn redis.Redis
}

type repositories struct {
	ruleDiller *rule_diller_storage.Storage
}

type consumers struct {
	ruleDiller *rule_diller_consumer.Consumer
}

type application struct {
	connections  connections
	repositories repositories
	provider     *provider.Provider
	service      *rule_diller_service.Implementation

	cfg Config
	wg  *sync.WaitGroup
}

func newApp(ctx context.Context, cfg *Config) *application {
	a := application{
		cfg: *cfg,
		wg:  &sync.WaitGroup{},
	}

	a.initConnections(ctx)
	a.initRepos(ctx)
	a.initProvider()
	a.initConsumer(ctx)
	a.initService()

	return &a
}

func (a *application) initConnections(ctx context.Context) {
	var err error
	a.connections.redisConn, err = redis.NewRedis(ctx, a.cfg.Redis.Dsn)
	if err != nil {
		logger.Error("init connect to database", zap.Error(err))
	}
}

func (a *application) initRepos(ctx context.Context) {
	ruleDiller, err := rule_diller_storage.New(ctx, a.connections.redisConn)
	if err != nil {
		logger.Fatal("init rule-diller repo", zap.Error(err))
	}

	a.repositories = repositories{
		ruleDiller: ruleDiller,
	}
}

func (a *application) initProvider() {
	a.provider = provider.NewProvider()
}

func (a *application) initConsumer(ctx context.Context) {
}

func (a *application) initService() {
	a.service = rule_diller_service.NewService(a.provider)
}

func (a *application) Run(ctx context.Context) error {
	server.StartRuleDiller(ctx, a.service, a.wg, a.cfg.Service.GrpcAddress)
	server.InitRuleDillerSwagger(ctx, a.wg, a.cfg.Service.SwaggerAddress, a.cfg.Service.SwaggerHost, a.cfg.Service.GrpcAddress)

	return nil
}

func (a *application) Close() {
	a.connections.redisConn.Close()
}
