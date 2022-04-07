package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var cfg ServerConfig

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	addr := pflag.StringP("addr", "a", cfg.Addr, "Server address")
	restore := pflag.BoolP("restore", "r", cfg.Restore, "Restore data on start up")
	storeInterval := pflag.DurationP("store-interval", "i", cfg.StoreInterval, "Store interval. 0 enables sync mode")
	storeFile := pflag.StringP("file", "f", cfg.StoreFile, "Path for file storage. Empty value disables file storage")

	pflag.Parse()

	cfg.Addr = *addr
	cfg.Restore = *restore
	cfg.StoreInterval = *storeInterval
	cfg.StoreFile = *storeFile

	s, err := NewMetricsServer(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)

	go s.Run(errChan)
	defer s.Stop()

	signal.Ignore(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-sigChan:
		log.Println("Terminating...")
		os.Exit(0)
	case err := <-errChan:
		log.Println("Server error: " + err.Error())
		os.Exit(1)
	}
}
