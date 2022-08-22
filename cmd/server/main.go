package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/vleukhin/prom-light/internal"
)

var buildVersion = "N/A"
var buildDate = "N/A"
var buildCommit = "N/A"

func main() {
	printIntro()
	cfg := &internal.ServerConfig{}
	if err := cfg.Parse(); err != nil {
		log.Fatal().Msg(err.Error())
	}

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	zerolog.SetGlobalLevel(logLevel)

	server, err := internal.NewMetricsServer(cfg)
	if err != nil {
		log.Fatal().Msg(err.Error())
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
		log.Info().Msg("Terminating...")
	case err := <-errChan:
		log.Error().Msg("Server error: " + err.Error())
	}
}

func printIntro() {
	fmt.Println("PromLight Server")
	fmt.Println("----------------")
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
	fmt.Println("----------------")
}
