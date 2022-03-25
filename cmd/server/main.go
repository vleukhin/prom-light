package main

import (
	"fmt"
	"github.com/vleukhin/prom-light/cmd/server/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := server.ServerConfig{
		Addr: "0.0.0.0",
		Port: 8080,
	}

	s := server.NewMetricsServer(cfg)
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
