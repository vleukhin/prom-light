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
	var cfg ServerConfig

	addr := pflag.StringP("addr", "a", "localhost:8080", "Server address")
	restore := pflag.BoolP("restore", "r", true, "Restore data on start up")
	storeInterval := pflag.DurationP("store-interval", "i", 300*time.Second, "Store interval. 0 enables sync mode")
	storeFile := pflag.StringP("file", "f", "/tmp/devops-metrics-db.json", "Path for file storage. Empty value disables file storage")

	pflag.Parse()

	cfg.Addr = *addr
	cfg.Restore = *restore
	cfg.StoreInterval = *storeInterval
	cfg.StoreFile = *storeFile

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
