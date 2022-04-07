package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
	"log"
	"time"
)

type CollectorConfig struct {
	PollInterval   time.Duration `env:"POLL_INTERVAL"   envDefault:"2s"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	ReportTimeout  time.Duration `env:"REPORT_TIMEOUT"  envDefault:"1s"`
	ServerAddr     string        `env:"ADDRESS"         envDefault:"localhost:8080"`
}

func (cfg *CollectorConfig) Init() error {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	serverAddr := pflag.StringP("addr", "a", cfg.ServerAddr, "Server address")
	pollInterval := pflag.DurationP("poll-interval", "p", cfg.PollInterval, "Poll interval")
	reportInterval := pflag.DurationP("report-interval", "r", cfg.ReportInterval, "Report interval")
	reportTimeout := pflag.DurationP("report-timeout", "t", cfg.ReportTimeout, "Report timeout")

	pflag.Parse()

	cfg.ServerAddr = *serverAddr
	cfg.PollInterval = *pollInterval
	cfg.ReportInterval = *reportInterval
	cfg.ReportTimeout = *reportTimeout

	return nil
}
