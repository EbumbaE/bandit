package run

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	bandit_indexer_client "github.com/EbumbaE/bandit/pkg/genproto/bandit-indexer/api"
	rule_admin_client "github.com/EbumbaE/bandit/pkg/genproto/rule-admin/api"
	"github.com/EbumbaE/bandit/pkg/kafka"
	"github.com/EbumbaE/bandit/pkg/logger"
	"github.com/EbumbaE/bandit/pkg/redis"

	rule_diller_service "github.com/EbumbaE/bandit/services/rule-diller/app"
	client_wrapper "github.com/EbumbaE/bandit/services/rule-diller/internal/client"
	"github.com/EbumbaE/bandit/services/rule-diller/internal/consumer"
	"github.com/EbumbaE/bandit/services/rule-diller/internal/provider"
	rule_diller_storage "github.com/EbumbaE/bandit/services/rule-diller/internal/storage"
	"github.com/EbumbaE/bandit/services/rule-diller/server"
)

type clients struct {
	indexerWrapper *client_wrapper.IndexerWrapper
	adminWrapper   *client_wrapper.AdminWrapper
}

type connections struct {
	redisConn *redis.Client
}

type repositories struct {
	ruleDiller *rule_diller_storage.Storage
}

type consumers struct {
	ruleDiller kafka.KafkaConsumer
}

type application struct {
	clients      clients
	connections  connections
	repositories repositories
	provider     *provider.Provider
	consumers    consumers
	service      *rule_diller_service.Implementation

	cfg Config
	wg  *sync.WaitGroup
}

func newApp(ctx context.Context, cfg *Config) *application {
	a := application{
		cfg: *cfg,
		wg:  &sync.WaitGroup{},
	}

	a.initClients(ctx)
	a.initConnections(ctx)
	a.initRepos()
	a.initProvider()
	a.initConsumer(ctx)
	a.initService()

	return &a
}

func (a *application) initClients(ctx context.Context) {
	conn, err := grpc.DialContext(ctx, a.cfg.Service.BanditIndexerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("connect to bandit-indexer", zap.Error(err))
	}

	a.clients.indexerWrapper = client_wrapper.NewIndexerWrapper(bandit_indexer_client.NewBanditIndexerServiceClient(conn))

	conn, err = grpc.DialContext(ctx, a.cfg.Service.RuleAdminAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("connect to rule-admin", zap.Error(err))
	}

	a.clients.adminWrapper = client_wrapper.NewAdminWrapper(rule_admin_client.NewRuleAdminServiceClient(conn))
}

func (a *application) initConnections(ctx context.Context) {
	var err error
	a.connections.redisConn, err = redis.NewRedis(ctx, a.cfg.Redis.Dsn)
	if err != nil {
		logger.Error("init connect to database", zap.Error(err))
	}
}

func (a *application) initRepos() {
	ruleDiller := rule_diller_storage.NewStorage(a.connections.redisConn)

	a.repositories = repositories{
		ruleDiller: ruleDiller,
	}
}

func (a *application) initProvider() {
	a.provider = provider.NewProvider(a.repositories.ruleDiller)
}

func (a *application) initConsumer(ctx context.Context) {
	handler := consumer.NewConsumer(a.clients.indexerWrapper, a.clients.adminWrapper, a.repositories.ruleDiller)

	consumer, err := kafka.NewKafkaConsumer(ctx, a.cfg.Kafka.Brokers, a.cfg.Kafka.Topic, handler.Handle)
	if err != nil {
		logger.Fatal("init rule-diller indexer consumer", zap.Error(err))
	}

	go consumer.Consume(ctx)

	a.consumers.ruleDiller = consumer
}

func (a *application) initService() {
	a.service = rule_diller_service.NewService(a.provider)
}

func (a *application) Run(ctx context.Context, swaggerPath string) error {
	server.StartRuleDiller(ctx, a.service, a.wg, a.cfg.Service.GrpcAddress)
	server.InitRuleDillerSwagger(ctx, a.wg, swaggerPath, a.cfg.Service.SwaggerAddress, a.cfg.Service.SwaggerHost, a.cfg.Service.GrpcAddress)

	return nil
}

func (a *application) Close() {
	a.connections.redisConn.Close()
	a.consumers.ruleDiller.Close()
}
