package main

import (
	"fmt"
	"github.com/vleukhin/prom-light/internal"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := internal.CollectorConfig{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		ReportTimeout:  1 * time.Second,
		ServerHost:     "127.0.0.1",
		ServerPort:     8080,
	}
	collector := internal.NewCollector(cfg)

	go collector.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-sigChan:
		fmt.Println("Terminating...")
		os.Exit(0)
	}
}
