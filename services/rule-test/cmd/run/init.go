package run

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	rule_admin_client "github.com/EbumbaE/bandit/pkg/genproto/rule-admin/api"
	rule_diller_client "github.com/EbumbaE/bandit/pkg/genproto/rule-diller/api"
	"github.com/EbumbaE/bandit/pkg/kafka"
	"github.com/EbumbaE/bandit/pkg/logger"
	"github.com/EbumbaE/bandit/pkg/psql"

	rule_test_service "github.com/EbumbaE/bandit/services/rule-test/app"
	wrapper "github.com/EbumbaE/bandit/services/rule-test/internal/client"
	"github.com/EbumbaE/bandit/services/rule-test/internal/metrics"
	"github.com/EbumbaE/bandit/services/rule-test/internal/notifier"
	"github.com/EbumbaE/bandit/services/rule-test/internal/provider"
	"github.com/EbumbaE/bandit/services/rule-test/server"
)

type producers struct {
	ruleTest kafka.SyncProducer
}

type notifiers struct {
	ruleTest *notifier.Notifier
}

type connections struct {
	db psql.Database
}

type clients struct {
	ruleDiller *wrapper.RuleDillerWrapper
	ruleAdmin  *wrapper.RuleAdminWrapper
}

type application struct {
	producers producers
	notifiers notifiers
	clients   clients
	provider  *provider.Provider
	service   *rule_test_service.Implementation

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
	a.initProvider()
	a.initService()

	return &a
}

func (a *application) initProducers(ctx context.Context) {
	producer, err := kafka.NewSyncProducer(ctx, a.cfg.Kafka.Topic, a.cfg.Kafka.Brokers)
	if err != nil {
		logger.Fatal("init rule-test producer", zap.Error(err))
	}

	a.producers.ruleTest = producer
	a.notifiers.ruleTest = notifier.NewNotifier(producer)
}

func (a *application) initClients(ctx context.Context) {
	{
		conn, err := grpc.DialContext(ctx, a.cfg.Service.RuleDillerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Fatal("connect to rule-diller", zap.Error(err))
		}

		a.clients.ruleDiller = wrapper.NewRuleDillerWrapper(rule_diller_client.NewRuleDillerServiceClient(conn))
	}
	{
		conn, err := grpc.DialContext(ctx, a.cfg.Service.RuleAdminAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Fatal("connect to rule-admin", zap.Error(err))
		}

		a.clients.ruleAdmin = wrapper.NewRuleAdminWrapper(rule_admin_client.NewRuleAdminServiceClient(conn))
	}
}

func (a *application) initProvider() {
	a.provider = provider.NewProvider(a.clients.ruleDiller, a.clients.ruleAdmin, a.notifiers.ruleTest)
}

func (a *application) initService() {
	a.service = rule_test_service.NewService(a.provider)
}

func (a *application) Run(ctx context.Context, swaggerPath string) error {
	metrics.StartMetricsServer(ctx, a.cfg.Prometheus.Host)

	server.StartRuleTest(ctx, a.service, a.wg, a.cfg.Service.GrpcAddress)
	server.StartRuleTestSwagger(ctx, a.wg, swaggerPath, a.cfg.Service.SwaggerAddress, a.cfg.Service.SwaggerHost, a.cfg.Service.GrpcAddress)

	return nil
}

func (a *application) Close(ctx context.Context) {
	metrics.StopeMetricsServer(ctx)
	a.producers.ruleTest.Close()
}
