package pollers

import (
	"testing"
)

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
