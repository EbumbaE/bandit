package run

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/EbumbaE/bandit/pkg/clickhouse"
	"github.com/EbumbaE/bandit/pkg/kafka"
	"github.com/EbumbaE/bandit/pkg/logger"
	"github.com/EbumbaE/bandit/pkg/psql"

	"github.com/EbumbaE/bandit/services/rule-analytic/internal/consumer"
	"github.com/EbumbaE/bandit/services/rule-analytic/internal/notifier"
	"github.com/EbumbaE/bandit/services/rule-analytic/internal/storage"
)

type producers struct {
	ruleAdmin kafka.SyncProducer
}

type notifiers struct {
	ruleAdmin *notifier.Notifier
}

type connections struct {
	psqlDB  psql.Database
	clickDB clickhouse.Database
}

type repositories struct {
	ruleAnalytic *storage.Storage
}

type consumers struct {
	externalEvent kafka.KafkaConsumer
}

type application struct {
	producers    producers
	notifiers    notifiers
	connections  connections
	repositories repositories
	consumers    consumers

	cfg Config
	wg  *sync.WaitGroup
}

func newApp(ctx context.Context, cfg *Config) *application {
	a := application{
		cfg: *cfg,
		wg:  &sync.WaitGroup{},
	}

	a.initConnections(ctx)
	a.initProducers(ctx)
	a.initRepos(ctx)
	a.initConsumer(ctx)

	return &a
}

func (a *application) initConnections(ctx context.Context) {
	var err error
	a.connections.psqlDB, err = psql.NewDatabase(ctx, a.cfg.Postgres.Dsn)
	if err != nil {
		logger.Error("init connect to psql database", zap.Error(err))
	}

	a.connections.clickDB, err = clickhouse.NewDatabase(ctx, a.cfg.Postgres.Dsn)
	if err != nil {
		logger.Error("init connect to click database", zap.Error(err))
	}
}

func (a *application) initProducers(ctx context.Context) {
	producer, err := kafka.NewSyncProducer(ctx, a.cfg.Kafka.InternalTopic, a.cfg.Kafka.Brokers)
	if err != nil {
		logger.Fatal("init rule-analytic producer", zap.Error(err))
	}

	a.producers.ruleAdmin = producer
	a.notifiers.ruleAdmin = notifier.NewNotifier(producer)
}

func (a *application) initRepos(ctx context.Context) {
	ruleAnalytic, err := storage.New(ctx, a.connections.psqlDB, a.connections.clickDB)
	if err != nil {
		logger.Fatal("init rule-analytic producer", zap.Error(err))
	}

	a.repositories = repositories{
		ruleAnalytic: ruleAnalytic,
	}
}

func (a *application) initConsumer(ctx context.Context) {
	handler := consumer.NewConsumer(a.repositories.ruleAnalytic, a.notifiers.ruleAdmin)

	consumer, err := kafka.NewKafkaConsumer(ctx, a.cfg.Kafka.Brokers, a.cfg.Kafka.ExternalTopic, handler.Handle)
	if err != nil {
		logger.Fatal("init rule-analytic indexer consumer", zap.Error(err))
	}

	a.consumers.externalEvent = consumer
}

func (a *application) Run(ctx context.Context) error {
	go a.consumers.externalEvent.Consume(ctx)

	return nil
}

func (a *application) Close() {
	a.connections.psqlDB.Close()
	a.producers.ruleAdmin.Close()

	if err := a.connections.clickDB.Close(); err != nil {
		logger.Error("rule-analytic failed to close clickDB", zap.Error(err))
	}
}
