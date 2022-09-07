package config

import (
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
)

// AgentConfig описывает конфиг агента
type AgentConfig struct {
	ServerAddr     string   `env:"ADDRESS" json:"address"`
	PollInterval   Duration `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval Duration `env:"REPORT_INTERVAL" json:"report_interval"`
	ReportTimeout  Duration `env:"REPORT_TIMEOUT" json:"report_timeout"`
	Key            string   `env:"KEY" json:"hash_key"`
	BatchMode      bool     `env:"BATCH_MODE" json:"batch_mode"`
	LogLevel       string   `env:"LOG_LEVEL" json:"log_level"`
	CryptoKey      string   `env:"CRYPTO_KEY" json:"crypto_key"`
}

func (cfg *AgentConfig) Parse() error {
	err := fillConfigFromFile(cfg)
	if err != nil {
		return err
	}
	serverAddr := pflag.StringP("addr", "a", "localhost:8080", "Server address")
	pollInterval := pflag.DurationP("poll-interval", "p", 2*time.Second, "Poll interval")
	reportInterval := pflag.DurationP("report-interval", "r", 10*time.Second, "Report interval")
	reportTimeout := pflag.DurationP("report-timeout", "t", 1*time.Second, "Report timeout")
	key := pflag.StringP("key", "k", "", "Secret key for signing data")
	batch := pflag.BoolP("batch", "b", true, "Report metrics in batches")
	logLevel := pflag.StringP("log-level", "l", "info", "Setup log level")
	cryptoKey := pflag.StringP("crypto-key", "e", "", "Path to public key")

	pflag.Parse()

	cfg.ServerAddr = *serverAddr
	cfg.PollInterval = Duration{*pollInterval}
	cfg.ReportInterval = Duration{*reportInterval}
	cfg.ReportTimeout = Duration{*reportTimeout}
	cfg.Key = *key
	cfg.BatchMode = *batch
	cfg.LogLevel = *logLevel
	cfg.CryptoKey = *cryptoKey

	err = env.Parse(cfg)
	if err != nil {
		return err
	}

	return nil
}
