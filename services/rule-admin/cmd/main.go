package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v2"

	"github.com/EbumbaE/bandit/pkg/database"
	rule_diller_client "github.com/EbumbaE/bandit/pkg/genproto/rule-diller/api"
	memcache "github.com/EbumbaE/bandit/pkg/inmemory_cache"
	"github.com/EbumbaE/bandit/pkg/logger"

	rule_admin_service "github.com/EbumbaE/bandit/services/rule-admin/app"
	"github.com/EbumbaE/bandit/services/rule-admin/internal/cache"
	rule_diller_wrapper "github.com/EbumbaE/bandit/services/rule-admin/internal/client"
	"github.com/EbumbaE/bandit/services/rule-admin/internal/provider"
	"github.com/EbumbaE/bandit/services/rule-admin/server"
)

type connections struct {
	memcached *memcache.Cache[string, []byte]
	database  database.Database
}

type repositories struct {
	cache *cache.Cache
}

type clients struct {
	ruleDiller *rule_diller_wrapper.RuleDillerWrapper
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

	a.initClients(ctx)
	a.initConnections(ctx)
	a.initRepos(ctx)
	a.initProvider()
	a.initService()

	return &a
}

func (a *application) initClients(ctx context.Context) {
	conn, err := grpc.DialContext(ctx, a.cfg.Service.RuleDillerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("connect to account-service", zap.Error(err))
	}

	a.clients.ruleDiller = rule_diller_wrapper.NewRuleDillerWrapper(rule_diller_client.NewRuleDillerServiceClient(conn))
}

func (a *application) initConnections(ctx context.Context) {
	a.connections.memcached = memcache.NewCache[string, []byte]()

	var err error
	a.connections.database, err = database.NewDatabase(ctx, a.cfg.Postgres.Dsn)
	if err != nil {
		logger.Error("init connect to database", zap.Error(err))
	}
}

func (a *application) initRepos(ctx context.Context) {
	a.repositories = repositories{
		cache: cache.NewCache(a.connections.memcached),
	}
}

func (a *application) initProvider() {
	a.provider = provider.NewProvider()
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
	a.connections.database.Close()
}

func main() {
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
	Service  TokenService `yaml:"service"`
	Postgres Postgres     `yaml:"postgres"`
}

type TokenService struct {
	SwaggerAddress    string        `yaml:"swagger_address"`
	SwaggerHost       string        `yaml:"swagger_host"`
	GrpcAddress       string        `yaml:"rule_admin_address"`
	RuleDillerAddress string        `yaml:"rule_diller_address"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout"`
}

type Postgres struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	DBName   string `yaml:"dbname"`
	Password string
	Dsn      string
}
