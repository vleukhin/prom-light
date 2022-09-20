package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/vleukhin/prom-light/internal/agent"
	"github.com/vleukhin/prom-light/internal/config"
)

var buildVersion = "N/A"
var buildDate = "N/A"
var buildCommit = "N/A"

func main() {
	printIntro()
	cfg := &config.AgentConfig{}
	if err := cfg.Parse(); err != nil {
		log.Fatal().Msg(err.Error())
	}

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	zerolog.SetGlobalLevel(logLevel)

	app, err := agent.NewApp(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create app")
	}
	mainCtx, cancel := context.WithCancel(context.Background())
	go app.Start(mainCtx, cancel)
	errChan := make(chan error)
	go func() {
		errChan <- http.ListenAndServe("localhost:8888", nil)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-sigChan:
		cancel()
		log.Info().Msg("Terminating...")
		ctx, stopCancel := context.WithTimeout(context.Background(), time.Second*5)
		app.Stop(ctx)
		stopCancel()
		return
	case err := <-errChan:
		log.Error().Msg("Server error: " + err.Error())
	case <-mainCtx.Done():
		log.Info().Msg("Application stopped by app...")
	}
}

func printIntro() {
	fmt.Println("PromLight Agent")
	fmt.Println("----------------")
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
	fmt.Println("----------------")
}
