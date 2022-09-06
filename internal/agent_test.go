package internal

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgent_Start(t *testing.T) {
	agent, err := NewAgent(&AgentConfig{
		PollInterval:   50 * time.Millisecond,
		ReportInterval: 50 * time.Millisecond,
	})

	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	go agent.Start(ctx, cancel)
	time.Sleep(100 * time.Millisecond)
	cancel()
}
