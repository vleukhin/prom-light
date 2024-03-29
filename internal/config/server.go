package config

import (
	"net"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
)

// ServerConfig описывает конфиг сервера
type ServerConfig struct {
	Addr          string    `env:"ADDRESS" json:"address"`
	Restore       bool      `env:"RESTORE" json:"restore"`
	StoreFile     string    `env:"STORE_FILE" json:"store_file"`
	StoreInterval Duration  `env:"STORE_INTERVAL" json:"store_interval"`
	Key           string    `env:"KEY" json:"hash_key"`
	DSN           string    `env:"DATABASE_DSN" json:"database_dsn"`
	DBConnTimeout Duration  `env:"DB_CONN_TIMEOUT" envDefault:"5s" json:"db_conn_timeout"`
	LogLevel      string    `env:"LOG_LEVEL" json:"log_level"`
	CryptoKey     string    `env:"CRYPTO_KEY" json:"crypto_key"`
	TrustedSubnet net.IPNet `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	Protocol      string    `env:"PROTOCOL" json:"protocol"`
}

func (cfg *ServerConfig) Parse() error {
	err := fillConfigFromFile(cfg)
	if err != nil {
		return err
	}
	addr := pflag.StringP("addr", "a", "localhost:8080", "Server address")
	restore := pflag.BoolP("restore", "r", true, "Restore data on start up")
	storeInterval := pflag.DurationP("store-interval", "i", 1*time.Minute, "Store interval. 0 enables sync mode")
	storeFile := pflag.StringP("file", "f", "/tmp/devops-metrics-db.json", "Path for file storage. Empty value disables file storage")
	key := pflag.StringP("key", "k", "", "Secret key for signing data")
	dsn := pflag.StringP("database-dsn", "d", "", "Database connection string")
	logLevel := pflag.StringP("log-level", "l", "info", "Setup log level")
	cryptoKey := pflag.StringP("crypto-key", "e", "", "Path to private key")
	trustedSubnet := pflag.IPNetP("trusted-subnet", "t", net.IPNet{}, "CIDR for trusted subnet")
	proto := pflag.StringP("protocol", "p", "http", "Server protocol (http or grpc")

	pflag.Parse()

	cfg.Addr = *addr
	cfg.Restore = *restore
	cfg.StoreInterval = Duration{*storeInterval}
	cfg.StoreFile = *storeFile
	cfg.Key = *key
	cfg.DSN = *dsn
	cfg.LogLevel = *logLevel
	cfg.CryptoKey = *cryptoKey
	cfg.TrustedSubnet = *trustedSubnet
	cfg.Protocol = *proto

	err = env.ParseWithFuncs(cfg, parseFuncs())
	if err != nil {
		return err
	}

	return nil
}
