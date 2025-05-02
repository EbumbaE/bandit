package run

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	rule_admin_client "github.com/EbumbaE/bandit/pkg/genproto/rule-admin/api"
	"github.com/EbumbaE/bandit/pkg/kafka"
	"github.com/EbumbaE/bandit/pkg/logger"
	"github.com/EbumbaE/bandit/pkg/psql"

	bandit_indexer_service "github.com/EbumbaE/bandit/services/bandit-indexer/app"
	client_wrapper "github.com/EbumbaE/bandit/services/bandit-indexer/internal/client"
	indexer_consumer "github.com/EbumbaE/bandit/services/bandit-indexer/internal/consumer"
	"github.com/EbumbaE/bandit/services/bandit-indexer/internal/notifier"
	"github.com/EbumbaE/bandit/services/bandit-indexer/internal/provider"
	indexer_storage "github.com/EbumbaE/bandit/services/bandit-indexer/internal/storage"
	"github.com/EbumbaE/bandit/services/bandit-indexer/server"
)

type producers struct {
	banditIndexer kafka.SyncProducer
}

type notifiers struct {
	banditIndexer *notifier.Notifier
}

type clients struct {
	adminWrapper *client_wrapper.AdminWrapper
}

type connections struct {
	db psql.Database
}

type repositories struct {
	banditIndexer *indexer_storage.Storage
}

type consumers struct {
	ruleAdminEvent kafka.KafkaConsumer
}

type application struct {
	producers    producers
	notifiers    notifiers
	clients      clients
	connections  connections
	repositories repositories
	provider     *provider.Provider
	service      *bandit_indexer_service.Implementation
	consumers    consumers

	cfg Config
	wg  *sync.WaitGroup
}

func newApp(ctx context.Context, cfg *Config) *application {
	a := application{
		cfg: *cfg,
		wg:  &sync.WaitGroup{},
	}

	a.initClients(ctx)
	a.initProducers(ctx)
	a.initConnections(ctx)
	a.initRepos(ctx)
	a.initProvider()
	a.initService()
	a.initConsumer(ctx)

	return &a
}

func (a *application) initClients(ctx context.Context) {
	conn, err := grpc.DialContext(ctx, a.cfg.Service.RuleAdminAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("connect to bandit-indexer", zap.Error(err))
	}

	a.clients.adminWrapper = client_wrapper.NewAdminWrapper(rule_admin_client.NewRuleAdminServiceClient(conn))
}

func (a *application) initProducers(ctx context.Context) {
	producer, err := kafka.NewSyncProducer(ctx, a.cfg.Kafka.IndexerTopic, a.cfg.Kafka.Brokers)
	if err != nil {
		logger.Fatal("init bandit-indexer producer", zap.Error(err))
	}

	a.producers.banditIndexer = producer
	a.notifiers.banditIndexer = notifier.NewNotifier(producer)
}

func (a *application) initConnections(ctx context.Context) {
	var err error
	a.connections.db, err = psql.NewDatabase(ctx, a.cfg.Postgres.Dsn)
	if err != nil {
		logger.Error("init connect to database", zap.Error(err))
	}
}

func (a *application) initRepos(ctx context.Context) {
	banditIndexer, err := indexer_storage.New(ctx, a.connections.db)
	if err != nil {
		logger.Fatal("init bandit-indexer storage", zap.Error(err))
	}

	a.repositories = repositories{
		banditIndexer: banditIndexer,
	}
}

func (a *application) initProvider() {
	a.provider = provider.NewProvider(a.repositories.banditIndexer)
}

func (a *application) initService() {
	a.service = bandit_indexer_service.NewService(a.provider)
}

func (a *application) initConsumer(ctx context.Context) {
	admin := indexer_consumer.NewAdminConsumer(a.clients.adminWrapper, a.repositories.banditIndexer, a.notifiers.banditIndexer)
	consumer, err := kafka.NewKafkaConsumer(ctx, a.cfg.Kafka.Brokers, a.cfg.Kafka.AdminTopic, admin.Handle, nil)
	if err != nil {
		logger.Fatal("init bandit-indexer admin consumer", zap.Error(err))
	}

	go consumer.Consume(ctx)

	a.consumers.ruleAdminEvent = consumer

	analytic := indexer_consumer.NewAnalyticConsumer(a.provider, a.notifiers.banditIndexer)
	consumer, err = kafka.NewKafkaConsumer(ctx, a.cfg.Kafka.Brokers, a.cfg.Kafka.AnalyticTopic, analytic.Handle, nil)
	if err != nil {
		logger.Fatal("init bandit-indexer analytic consumer", zap.Error(err))
	}

	go consumer.Consume(ctx)

	a.consumers.ruleAdminEvent = consumer
}

func (a *application) Run(ctx context.Context, swaggerPath string) error {
	server.StarBanditIndexer(ctx, a.service, a.wg, a.cfg.Service.GrpcAddress)
	server.InitBanditIndexerSwagger(ctx, a.wg, swaggerPath, a.cfg.Service.SwaggerAddress, a.cfg.Service.SwaggerHost, a.cfg.Service.GrpcAddress)

	return nil
}

func (a *application) Close() {
	a.connections.db.Close()
}
