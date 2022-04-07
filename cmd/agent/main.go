package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := &CollectorConfig{}
	if err := cfg.Init(); err != nil {
		log.Fatal(err.Error())
	}
	collector := NewCollector(cfg)

	go collector.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigChan
	log.Println("Terminating...")
	collector.Stop()
	os.Exit(0)
}
