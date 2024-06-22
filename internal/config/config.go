package config

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/dmad1989/gophKeeper/tools/model/consts"
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
	var serverPort string
	pflag.StringVarP(&serverPort, "a", "a", "", "Port of the proto server")

	var dbConn string
	pflag.StringVarP(&dbConn, "d", "d", "", "Postgres DB DSN")
	if serverPort != "" {
		cfg.ServerPort = serverPort
	}
	if dbConn != "" {
		cfg.DBConn = dbConn
	}
}

type Config struct {
	ServerPort string `json:"server_port"`
	DBConn     string `json:"db_conn"`
}

func (c Config) String() string {
	return fmt.Sprintf(configFormat, c.ServerPort, c.DBConn)
}

func (c Config) GetDBConn() string {
	return c.DBConn
}

func (c Config) GetServerPort() string {
	return c.ServerPort
}
