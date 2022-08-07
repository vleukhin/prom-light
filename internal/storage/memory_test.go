package storage

import (
	"context"
	"testing"
)

func TestMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage()

	testStorage(storage, t)
	if err := storage.CleanUp(context.Background()); err != nil {
		panic(err)
	}
}
