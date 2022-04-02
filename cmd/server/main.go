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
	var cfg ServerConfig

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	s := NewMetricsServer(cfg)
	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)

	go s.Run(errChan)

	signal.Ignore(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-sigChan:
		fmt.Println("Terminating...")
		os.Exit(0)
	case err := <-errChan:
		fmt.Println("Server error: " + err.Error())
		os.Exit(1)
	}
}
