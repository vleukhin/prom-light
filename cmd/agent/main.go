package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
)

func main() {
	var cfg CollectorConfig

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
	fmt.Println("Terminating...")
	collector.Stop()
	os.Exit(0)
}
