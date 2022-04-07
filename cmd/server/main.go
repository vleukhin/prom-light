package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := &ServerConfig{}
	if err := cfg.Init(); err != nil {
		log.Fatal(err.Error())
	}

	server, err := NewMetricsServer(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)

	go server.Run(errChan)
	defer server.Stop()

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
