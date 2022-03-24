package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := ServerConfig{
		Addr: "0.0.0.0",
		Port: 8080,
	}

	server := newMetricsServer(cfg)
	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)

	go server.run(errChan)

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
