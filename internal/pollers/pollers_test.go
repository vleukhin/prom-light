package pollers

import (
	"testing"
)

func TestPollers(t *testing.T) {
	t.Run("ps", func(t *testing.T) {
		poller := PsPoller{}
		_, err := poller.Poll()
		if err != nil {
			t.Errorf("Got error from ps poller")
		}
	})
	t.Run("mstats", func(t *testing.T) {
		poller := MemStatsPoller{}
		_, err := poller.Poll()
		if err != nil {
			t.Errorf("Got error from mstats poller")
		}
	})
}

func BenchmarkPollers(b *testing.B) {
	mstatsPoller := MemStatsPoller{}
	gopsPoller := PsPoller{}

	b.ResetTimer()
	b.Run("mstats", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mstatsPoller.Poll()
		}
	})
	b.Run("gops", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			gopsPoller.Poll()
		}
	})
}
