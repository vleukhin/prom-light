package internal

import (
	"context"
	"testing"
	"time"
)

func TestAgent_Start(t *testing.T) {
	agent := NewAgent(&AgentConfig{
		PollInterval:   50 * time.Millisecond,
		ReportInterval: 50 * time.Millisecond,
	})

	ctx, cancel := context.WithCancel(context.Background())
	go agent.Start(ctx, cancel)
	time.Sleep(100 * time.Millisecond)
	cancel()
}
