package internal

import (
	"github.com/caarlos0/env/v6"
	"time"

	"github.com/spf13/pflag"
)

// ServerConfig описывает конфиг сервера
type ServerConfig struct {
	Addr          string        `env:"ADDRESS"`
	Restore       bool          `env:"RESTORE"`
	StoreFile     string        `env:"STORE_FILE"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	Key           string        `env:"KEY"`
	DSN           string        `env:"DATABASE_DSN"`
	DBConnTimeout time.Duration `env:"DB_CONN_TIMEOUT" envDefault:"5s"`
	LogLevel      string        `env:"LOG_LEVEL"`
	CryptoKey     string        `env:"CRYPTO_KEY"`
}

// AgentConfig описывает конфиг агента
type AgentConfig struct {
	ServerAddr     string        `env:"ADDRESS"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	ReportTimeout  time.Duration `env:"REPORT_TIMEOUT"`
	Key            string        `env:"KEY"`
	BatchMode      bool          `env:"BATCH_MODE"`
	LogLevel       string        `env:"LOG_LEVEL"`
	CryptoKey      string        `env:"CRYPTO_KEY"`
}

func (cfg *ServerConfig) Parse() error {
	addr := pflag.StringP("addr", "a", "localhost:8080", "Server address")
	restore := pflag.BoolP("restore", "r", true, "Restore data on start up")
	storeInterval := pflag.DurationP("store-interval", "i", 1*time.Minute, "Store interval. 0 enables sync mode")
	storeFile := pflag.StringP("file", "f", "/tmp/devops-metrics-db.json", "Path for file storage. Empty value disables file storage")
	key := pflag.StringP("key", "k", "", "Secret key for signing data")
	dsn := pflag.StringP("database-dsn", "d", "", "Database connection string")
	logLevel := pflag.StringP("log-level", "l", "info", "Setup log level")
	cryptoKey := pflag.StringP("crypto-key", "c", "", "Path to private key")

	pflag.Parse()

	cfg.Addr = *addr
	cfg.Restore = *restore
	cfg.StoreInterval = *storeInterval
	cfg.StoreFile = *storeFile
	cfg.Key = *key
	cfg.DSN = *dsn
	cfg.LogLevel = *logLevel
	cfg.CryptoKey = *cryptoKey

	err := env.Parse(cfg)
	if err != nil {
		return err
	}

	return nil
}

func (cfg *AgentConfig) Parse() error {
	serverAddr := pflag.StringP("addr", "a", "localhost:8080", "Server address")
	pollInterval := pflag.DurationP("poll-interval", "p", 2*time.Second, "Poll interval")
	reportInterval := pflag.DurationP("report-interval", "r", 10*time.Second, "Report interval")
	reportTimeout := pflag.DurationP("report-timeout", "t", 1*time.Second, "Report timeout")
	key := pflag.StringP("key", "k", "", "Secret key for signing data")
	batch := pflag.BoolP("batch", "b", true, "Report metrics in batches")
	logLevel := pflag.StringP("log-level", "l", "info", "Setup log level")
	cryptoKey := pflag.StringP("crypto-key", "c", "", "Path to public key")

	pflag.Parse()

	cfg.ServerAddr = *serverAddr
	cfg.PollInterval = *pollInterval
	cfg.ReportInterval = *reportInterval
	cfg.ReportTimeout = *reportTimeout
	cfg.Key = *key
	cfg.BatchMode = *batch
	cfg.LogLevel = *logLevel
	cfg.CryptoKey = *cryptoKey

	err := env.Parse(cfg)
	if err != nil {
		return err
	}

	return nil
}
