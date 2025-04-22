package run

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/EbumbaE/bandit/pkg/kafka"
	"github.com/EbumbaE/bandit/pkg/logger"
	"github.com/EbumbaE/bandit/pkg/psql"

	rule_admin_service "github.com/EbumbaE/bandit/services/bandit-indexer/app"
	rule_admin_wrapper "github.com/EbumbaE/bandit/services/bandit-indexer/internal/client"
	"github.com/EbumbaE/bandit/services/bandit-indexer/internal/provider"
	bandit_indexer_storage "github.com/EbumbaE/bandit/services/bandit-indexer/internal/storage"
	"github.com/EbumbaE/bandit/services/bandit-indexer/server"
)

type clients struct {
	adminWrapper *rule_admin_wrapper.AdminWrapper
}

type connections struct {
	db psql.Database
}

type repositories struct {
	banditIndexer *bandit_indexer_storage.Storage
}

type consumers struct {
	ruleAdminEvent kafka.KafkaConsumer
}

type producers struct {
	banditIndexerEvent kafka.SyncProducer
}

type application struct {
	clients      clients
	connections  connections
	repositories repositories
	provider     *provider.Provider
	service      *rule_admin_service.Implementation

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
	a.initService()

	return &a
}

func (a *application) initConnections(ctx context.Context) {
	var err error
	a.connections.db, err = psql.NewDatabase(ctx, a.cfg.Postgres.Dsn)
	if err != nil {
		logger.Error("init connect to database", zap.Error(err))
	}
}

func (a *application) initRepos(ctx context.Context) {
	banditIndexer, err := bandit_indexer_storage.New(ctx, a.connections.db)
	if err != nil {
		logger.Fatal("init rule admin repo", zap.Error(err))
	}

	a.repositories = repositories{
		banditIndexer: banditIndexer,
	}
}

func (a *application) initProvider() {
	a.provider = provider.NewProvider()
}

func (a *application) initService() {
	a.service = rule_admin_service.NewService(a.provider)
}

func (a *application) Run(ctx context.Context) error {
	server.StartRuleDiller(ctx, a.service, a.wg, a.cfg.Service.GrpcAddress)
	server.InitRuleDillerSwagger(ctx, a.wg, a.cfg.Service.SwaggerAddress, a.cfg.Service.SwaggerHost, a.cfg.Service.GrpcAddress)

	return nil
}

func (a *application) Close() {
	a.connections.db.Close()
}
