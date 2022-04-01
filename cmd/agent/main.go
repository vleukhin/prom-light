package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := CollectorConfig{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		ReportTimeout:  1 * time.Second,
		ServerHost:     "127.0.0.1",
		ServerPort:     8080,
	}
	collector := NewCollector(cfg)

	go collector.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigChan
	fmt.Println("Terminating...")
	collector.Stop()
	os.Exit(0)
}
