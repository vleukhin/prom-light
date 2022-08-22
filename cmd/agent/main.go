package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/vleukhin/prom-light/internal"
)

func main() {
	cfg := &internal.AgentConfig{}
	if err := cfg.Parse(); err != nil {
		log.Fatal().Msg(err.Error())
	}

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	zerolog.SetGlobalLevel(logLevel)

	agent := internal.NewAgent(cfg)
	mainCtx, cancel := context.WithCancel(context.Background())
	go agent.Start(mainCtx, cancel)
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
		agent.Stop()
		return
	case err := <-errChan:
		log.Error().Msg("Server error: " + err.Error())
	case <-mainCtx.Done():
		log.Info().Msg("Application stopped by agent...")
	}
}
