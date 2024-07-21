package config

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/json"
	"encoding/pem"
	"errors"
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
	//go:embed privatekey.pem
	privateKeyBytes []byte
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
		return nil, fmt.Errorf("config.newConfig: reading config file, %w", err)
	}
	if err := cfg.parseFlags(); err != nil {
		return nil, fmt.Errorf("config.newConfig: %w", err)
	}
	log.Debug("read config", zap.Stringer("config", cfg))
	return &cfg, nil
}

func (cfg *Config) parseFlags() (err error) {
	sAddress := os.Getenv("GOPHKEEPER_SERVER_ADDRESS")
	if sAddress == "" {
		pflag.StringVarP(&sAddress, "a", "a", "", "Port of the proto server")
	}

	db := os.Getenv("GOPHKEEPER_DB_DSN")
	if db == "" {
		pflag.StringVarP(&db, "d", "d", "", "Postgres DB DSN")
	}

	keyPath := os.Getenv("GOPHKEEPER_PRIVATE_KEY_PATH")
	if keyPath == "" {
		pflag.StringVarP(&keyPath, "f", "f", "", "Path to RSA key to encode data. If not defined will use default")
	}

	pflag.Parse()

	if sAddress != "" {
		cfg.ServerAddress = sAddress
	}
	if db != "" {
		cfg.DBConn = db
	}

	cfg.privateKey, err = readPrivateKey(keyPath)
	if err != nil {
		return fmt.Errorf("cfg.parseFlags: %w", err)
	}
	return
}

type Config struct {
	ServerAddress  string `json:"server_address"`
	DBConn         string `json:"db_conn"`
	PrivateKeyPath string `json:"crypto_key_path"`
	privateKey     *rsa.PrivateKey
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
func (c Config) GetPrivateKey() *rsa.PrivateKey {
	return c.privateKey
}

func readPrivateKey(path string) (*rsa.PrivateKey, error) {
	var err error
	if path != "" {
		privateKeyBytes, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("readPrivateKey: os.ReadFile('%s'): %w", path, err)
		}
	}
	pemBlock, _ := pem.Decode(privateKeyBytes)
	if pemBlock == nil {
		return nil, errors.New("readPrivateKey: pem.Decode(privateKeyBytes): no PEM data found")
	}
	key, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Config.readPrivateKey: %w", err)
	}
	return key, nil
}
