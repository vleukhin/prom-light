package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
)

func main() {
	var cfg ServerConfig

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	s, err := NewMetricsServer(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)

	go s.Run(errChan)

	signal.Ignore(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-sigChan:
		log.Println("Terminating...")
		s.Stop()
		os.Exit(0)
	case err := <-errChan:
		log.Println("Server error: " + err.Error())
		os.Exit(1)
	}
}
