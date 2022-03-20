package main

import (
	"fmt"
	"github.com/vleukhin/prom-light/internal"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var pollInterval = 2 * time.Second
var reportInterval = 10 * time.Second

func main() {
	collector := internal.NewCollector(pollInterval, reportInterval)
	sigChan := make(chan os.Signal, 1)

	go collector.Start()

	signal.Ignore(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-sigChan:
		fmt.Println("Terminating...")
		os.Exit(0)
	}
}
