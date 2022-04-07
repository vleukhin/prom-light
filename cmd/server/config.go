package main

import (
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
)

type ServerConfig struct {
	Addr          string        `env:"ADDRESS"        envDefault:"localhost:8080"`
	Restore       bool          `env:"RESTORE"        envDefault:"true"`
	StoreFile     string        `env:"STORE_FILE"     envDefault:"/tmp/devops-metrics-db.json"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"1m"`
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

	pflag.Parse()

	cfg.Addr = *addr
	cfg.Restore = *restore
	cfg.StoreInterval = *storeInterval
	cfg.StoreFile = *storeFile

	return nil
}
