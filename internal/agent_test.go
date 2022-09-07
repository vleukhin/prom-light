package internal

import (
	"context"
	"testing"
	"time"

	"github.com/vleukhin/prom-light/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestAgent_Start(t *testing.T) {
	agent, err := NewAgent(&config.AgentConfig{
		PollInterval:   config.Duration{Duration: 50 * time.Millisecond},
		ReportInterval: config.Duration{Duration: 50 * time.Millisecond},
	})

	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	go agent.Start(ctx, cancel)
	time.Sleep(100 * time.Millisecond)
	cancel()
}
