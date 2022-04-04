package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/pflag"

	"github.com/caarlos0/env/v6"
)

func main() {
	var cfg CollectorConfig

	err := env.Parse(&cfg)
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

	collector := NewCollector(cfg)

	go collector.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigChan
	log.Println("Terminating...")
	collector.Stop()
	os.Exit(0)
}
