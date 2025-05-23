package run

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/EbumbaE/bandit/pkg/logger"
)

func Run() {
	configPath := flag.String("config", "", "config path")
	swaggerPath := flag.String("swagger", "", "swagger path")
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

	if err := app.Run(ctx, *swaggerPath); err != nil {
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
	Service RuleAdminService `yaml:"service"`
	Redis   Redis            `yaml:"redis"`
	Kafka   Kafka            `yaml:"kafka"`
}

type RuleAdminService struct {
	SwaggerAddress       string        `yaml:"swagger_address"`
	GrpcAddress          string        `yaml:"rule_diller_address"`
	SwaggerHost          string        `yaml:"swagger_host"`
	BanditIndexerAddress string        `yaml:"bandit_indexer_address"`
	RuleAdminAddress     string        `yaml:"rule_admin_address"`
	ConnectionTimeout    time.Duration `yaml:"connection_timeout"`
}

type Redis struct {
	Dsn string `yaml:"dsn"`
}

type Kafka struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}
