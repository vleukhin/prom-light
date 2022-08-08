package storage

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
)

func TestFileStorage(t *testing.T) {
	ctx := context.Background()
	storage, err := NewFileStorage("/tmp/metrics_tests", 5*time.Second, false)
	if err != nil {
		log.Fatal().Err(err)
	}
	testStorage(storage, t)
	if err := storage.CleanUp(ctx); err != nil {
		panic(err)
	}
	if err := storage.CleanUp(ctx); err != nil {
		panic(err)
	}
}
