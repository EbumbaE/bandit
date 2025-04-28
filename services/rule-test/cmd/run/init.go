package run

import (
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	rule_admin_client "github.com/EbumbaE/bandit/pkg/genproto/rule-admin/api"
	rule_diller_client "github.com/EbumbaE/bandit/pkg/genproto/rule-diller/api"
	"github.com/EbumbaE/bandit/pkg/kafka"
	"github.com/EbumbaE/bandit/pkg/logger"
	"github.com/EbumbaE/bandit/pkg/psql"

	wrapper "github.com/EbumbaE/bandit/services/rule-test/internal/client"
	"github.com/EbumbaE/bandit/services/rule-test/internal/metrics"
	"github.com/EbumbaE/bandit/services/rule-test/internal/notifier"
	"github.com/EbumbaE/bandit/services/rule-test/internal/provider"
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
			logger.Fatal("connect to rule-test", zap.Error(err))
		}

		a.clients.ruleDiller = wrapper.NewRuleDillerWrapper(rule_diller_client.NewRuleDillerServiceClient(conn))
	}
	{
		conn, err := grpc.DialContext(ctx, a.cfg.Service.RuleAdminAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Fatal("connect to rule-test", zap.Error(err))
		}

		a.clients.ruleAdmin = wrapper.NewRuleAdminWrapper(rule_admin_client.NewRuleAdminServiceClient(conn))
	}
}

func (a *application) initProvider() {
	a.provider = provider.NewProvider(a.clients.ruleDiller, a.clients.ruleAdmin, a.notifiers.ruleTest)
}

func (a *application) Run(ctx context.Context) error {
	metrics.InitMetricsServer(ctx)

	switch a.cfg.Test.Mode {
	case "work":
		return a.provider.DoEfficiencyTest(ctx, a.cfg.Test.CycleCount)
	case "load":
		return a.provider.DoLoadTest(ctx, a.cfg.Test.ParallelCount, a.cfg.Test.CycleCount)
	}

	return errors.New("undefined test mode")
}

func (a *application) Close(ctx context.Context) {
	metrics.StopeMetricsServer(ctx)
	a.producers.ruleTest.Close()
}
