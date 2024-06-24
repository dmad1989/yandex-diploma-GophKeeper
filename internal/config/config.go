package config

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

const configFormat = `Server port: "%s"  ;  DBConn:  "%s" ;`

var (
	//go:embed clientConfig.json
	clientConfig []byte
	//go:embed serverConfig.json
	serverConfig []byte
)

func NewServer(ctx context.Context) (*Config, error) {
	return newConfig(ctx, serverConfig)
}

func NewClient(ctx context.Context) (*Config, error) {
	return newConfig(ctx, clientConfig)
}

func newConfig(ctx context.Context, embConfig []byte) (*Config, error) {
	log := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("config")
	var cfg Config

	if err := json.Unmarshal(embConfig, &cfg); err != nil {
		return nil, fmt.Errorf("config.NewViper: reading config file, %w", err)
	}
	cfg.parseFlags()
	log.Debug("read config", zap.Stringer("config", cfg))
	return &cfg, nil
}

func (cfg *Config) parseFlags() {
	sAddress := os.Getenv("GOPHKEEPER_SERVER_ADDRESS")
	if sAddress == "" {
		pflag.StringVarP(&sAddress, "a", "a", "", "Port of the proto server")
	}

	db := os.Getenv("GOPHKEEPER_DB_DSN")
	if db == "" {
		pflag.StringVarP(&db, "d", "d", "", "Postgres DB DSN")
	}

	pflag.Parse()

	if sAddress != "" {
		cfg.ServerAddress = sAddress
	}
	if db != "" {
		cfg.DBConn = db
	}
}

type Config struct {
	ServerAddress string `json:"server_address"`
	DBConn        string `json:"db_conn"`
}

func (c Config) String() string {
	return fmt.Sprintf(configFormat, c.ServerAddress, c.DBConn)
}

func (c Config) GetDBConn() string {
	return c.DBConn
}

func (c Config) GetServerAddress() string {
	return c.ServerAddress
}
