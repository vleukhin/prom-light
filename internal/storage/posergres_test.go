package storage

import (
	"context"
	"testing"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
)

type testConfig struct {
	DSN string `env:"DATABASE_DSN_TEST" envDefault:""`
}

func TestPostgresStorage(t *testing.T) {
	var err error
	ctx := context.Background()
	cfg := testConfig{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal().Err(err)
	}

	if cfg.DSN == "" {
		return
	}
	db, err := NewPostgresStorage(cfg.DSN, time.Second*5)
	if err != nil {
		panic(err)
	}

	if err := db.Migrate(ctx); err != nil {
		panic(err)
	}
	testStorage(db, t)
	if err := db.CleanUp(ctx); err != nil {
		panic(err)
	}
}
