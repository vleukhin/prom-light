package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vleukhin/prom-light/internal"
)

func main() {
	cfg := &internal.ServerConfig{}
	if err := cfg.Init(); err != nil {
		log.Fatal(err.Error())
	}

	server, err := internal.NewMetricsServer(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)

	go server.Run(errChan)
	defer func(server internal.MetricsServer) {
		err := server.Stop()
		if err != nil {
			panic(err)
		}
	}(server)

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
