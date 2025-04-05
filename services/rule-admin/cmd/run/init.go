package run

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	rule_diller_client "github.com/EbumbaE/bandit/pkg/genproto/rule-diller/api"
	"github.com/EbumbaE/bandit/pkg/kafka"
	"github.com/EbumbaE/bandit/pkg/logger"
	"github.com/EbumbaE/bandit/pkg/psql"

	rule_admin_service "github.com/EbumbaE/bandit/services/rule-admin/app"
	rule_diller_wrapper "github.com/EbumbaE/bandit/services/rule-admin/internal/client"
	"github.com/EbumbaE/bandit/services/rule-admin/internal/notifier"
	"github.com/EbumbaE/bandit/services/rule-admin/internal/provider"
	rule_admin_storage "github.com/EbumbaE/bandit/services/rule-admin/internal/storage"
	"github.com/EbumbaE/bandit/services/rule-admin/server"
)

type connections struct {
	db psql.Database
}

type repositories struct {
	ruleAdmin *rule_admin_storage.Storage
}

type clients struct {
	ruleDiller *rule_diller_wrapper.RuleDillerWrapper
}

type producers struct {
	ruleAdminEvent kafka.SyncProducer
}

type notifiers struct {
	ruleAdminEvent *notifier.Notifier
}

type application struct {
	producers    producers
	notifiers    notifiers
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

	a.initProducers(ctx)
	a.initClients(ctx)
	a.initConnections(ctx)
	a.initRepos(ctx)
	a.initProvider()
	a.initService()

	return &a
}

func (a *application) initProducers(ctx context.Context) {
	producer, err := kafka.NewSyncProducer(ctx, a.cfg.Kafka.Topic, a.cfg.Kafka.Brokers)
	if err != nil {
		logger.Fatal("init rule-admin producer", zap.Error(err))
	}

	a.producers.ruleAdminEvent = producer
	a.notifiers.ruleAdminEvent = notifier.NewNotifier(producer)
}

func (a *application) initClients(ctx context.Context) {
	conn, err := grpc.DialContext(ctx, a.cfg.Service.RuleDillerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("connect to rule-admin", zap.Error(err))
	}

	a.clients.ruleDiller = rule_diller_wrapper.NewRuleDillerWrapper(rule_diller_client.NewRuleDillerServiceClient(conn))
}

func (a *application) initConnections(ctx context.Context) {
	var err error
	a.connections.db, err = psql.NewDatabase(ctx, a.cfg.Postgres.Dsn)
	if err != nil {
		logger.Error("init connect to database", zap.Error(err))
	}
}

func (a *application) initRepos(ctx context.Context) {
	ruleAdmin, err := rule_admin_storage.New(ctx, a.connections.db)
	if err != nil {
		logger.Fatal("init rule admin repo", zap.Error(err))
	}

	a.repositories = repositories{
		ruleAdmin: ruleAdmin,
	}
}

func (a *application) initProvider() {
	a.provider = provider.NewProvider(a.repositories.ruleAdmin, a.notifiers.ruleAdminEvent)
}

func (a *application) initService() {
	a.service = rule_admin_service.NewService(a.provider)
}

func (a *application) Run(ctx context.Context) error {
	server.StartRuleAdmin(ctx, a.service, a.wg, a.cfg.Service.GrpcAddress)
	server.InitRuleAdminSwagger(ctx, a.wg, a.cfg.Service.SwaggerAddress, a.cfg.Service.SwaggerHost, a.cfg.Service.GrpcAddress)

	return nil
}

func (a *application) Close() {
	a.connections.db.Close()
	a.producers.ruleAdminEvent.Close()
}
