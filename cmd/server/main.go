package main

import (
	"context"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/vleukhin/prom-light/internal/config"
	"github.com/vleukhin/prom-light/internal/server"
)

var buildVersion = "N/A"
var buildDate = "N/A"
var buildCommit = "N/A"

func main() {
	printIntro()
	cfg := &config.ServerConfig{}
	if err := cfg.Parse(); err != nil {
		log.Fatal().Msg(err.Error())
	}

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	zerolog.SetGlobalLevel(logLevel)

	app, err := server.NewApp(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create app")
	}

	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)

	go app.Run(errChan)
	defer func(server *server.App) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		err := server.Stop(ctx)
		cancel()
		if err != nil {
			log.Fatal().Err(err)
		}
	}(app)

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
