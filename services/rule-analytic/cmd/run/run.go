package run

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/EbumbaE/bandit/pkg/logger"
)

func Run() {
	configPath := flag.String("config", "", "config path")
	flag.Parse()

	config := readConfig(*configPath)

	ctx, cancel := context.WithCancel(context.Background())
	app := newApp(ctx, config)

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		sig := <-c
		logger.Info("received signal", zap.String("signal", sig.String()))
		cancel()
	}()

	defer app.Close()

	if err := app.Run(ctx); err != nil {
		logger.Fatal("can't run app", zap.Error(err))
	}

	app.wg.Wait()
}

func readConfig(configPath string) *Config {
	var cfg Config
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal("failed to read config from ", configPath, ":", err)
	}
	err = yaml.Unmarshal(bytes, &cfg)
	if err != nil {
		log.Fatal("failed to unmarshal config - ", err)
	}

	return &cfg
}

type Config struct {
	Postgres   Postgres   `yaml:"postgres"`
	ClickHouse ClickHouse `yaml:"clickhouse"`
	Kafka      Kafka      `yaml:"kafka"`
}

type Postgres struct {
	Dsn string `yaml:"dsn"`
}

type ClickHouse struct {
	Dsn string `yaml:"dsn"`
}

type Kafka struct {
	Brokers       []string `yaml:"brokers"`
	ExternalTopic string   `yaml:"external_topic"`
	InternalTopic string   `yaml:"internal_topic"`
}
