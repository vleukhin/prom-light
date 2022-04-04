package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/spf13/pflag"
)

func main() {
	var cfg CollectorConfig

	serverAddr := pflag.StringP("addr", "a", "localhost:8080", "Server address")
	pollInterval := pflag.DurationP("poll-interval", "p", 2*time.Second, "Poll interval")
	reportInterval := pflag.DurationP("report-interval", "r", 10*time.Second, "Report interval")
	reportTimeout := pflag.DurationP("report-timeout", "t", time.Second, "Report timeout")

	pflag.Parse()

	cfg.ServerAddr = *serverAddr
	cfg.PollInterval = *pollInterval
	cfg.ReportInterval = *reportInterval
	cfg.ReportTimeout = *reportTimeout

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

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
