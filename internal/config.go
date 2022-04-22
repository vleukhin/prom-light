package internal

import (
	"log"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
)

type ServerConfig struct {
	Addr          string        `env:"ADDRESS"        envDefault:"localhost:8080"`
	Restore       bool          `env:"RESTORE"        envDefault:"true"`
	StoreFile     string        `env:"STORE_FILE"     envDefault:"/tmp/devops-metrics-db.json"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"1m"`
	Key           string        `env:"KEY"`
}

type AgentConfig struct {
	PollInterval   time.Duration `env:"POLL_INTERVAL"   envDefault:"2s"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	ReportTimeout  time.Duration `env:"REPORT_TIMEOUT"  envDefault:"1s"`
	ServerAddr     string        `env:"ADDRESS"         envDefault:"localhost:8080"`
	Key            string        `env:"KEY"`
}

func (cfg *ServerConfig) Init() error {
	err := env.Parse(cfg)
	if err != nil {
		return err
	}

	addr := pflag.StringP("addr", "a", cfg.Addr, "Server address")
	restore := pflag.BoolP("restore", "r", cfg.Restore, "Restore data on start up")
	storeInterval := pflag.DurationP("store-interval", "i", cfg.StoreInterval, "Store interval. 0 enables sync mode")
	storeFile := pflag.StringP("file", "f", cfg.StoreFile, "Path for file storage. Empty value disables file storage")
	key := pflag.StringP("key", "k", cfg.Key, "Secret key for signing data")

	pflag.Parse()

	cfg.Addr = *addr
	cfg.Restore = *restore
	cfg.StoreInterval = *storeInterval
	cfg.StoreFile = *storeFile
	cfg.Key = *key

	return nil
}

func (cfg *AgentConfig) Init() error {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	serverAddr := pflag.StringP("addr", "a", cfg.ServerAddr, "Server address")
	pollInterval := pflag.DurationP("poll-interval", "p", cfg.PollInterval, "Poll interval")
	reportInterval := pflag.DurationP("report-interval", "r", cfg.ReportInterval, "Report interval")
	reportTimeout := pflag.DurationP("report-timeout", "t", cfg.ReportTimeout, "Report timeout")
	key := pflag.StringP("key", "k", cfg.Key, "Secret key for signing data")

	pflag.Parse()

	cfg.ServerAddr = *serverAddr
	cfg.PollInterval = *pollInterval
	cfg.ReportInterval = *reportInterval
	cfg.ReportTimeout = *reportTimeout
	cfg.Key = *key

	return nil
}
