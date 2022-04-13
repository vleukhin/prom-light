package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vleukhin/prom-light/internal"
)

func main() {
	cfg := &internal.AgentConfig{}
	if err := cfg.Init(); err != nil {
		log.Fatal(err.Error())
	}

	agent := internal.NewAgent(cfg)
	go agent.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigChan
	log.Println("Terminating...")
	agent.Stop()
	os.Exit(0)
}
