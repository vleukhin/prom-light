package internal

import (
	"math/rand"
	"runtime"
	"time"
)

type gauge float64
type counter int64

type Collector struct {
	gaugeMetrics   map[string]gauge
	counterMetrics map[string]counter
	ticker         *time.Ticker
}

func NewCollector(pollInterval, _ time.Duration) Collector {
	rand.Seed(time.Now().Unix())
	gauges := make(map[string]gauge)
	counters := make(map[string]counter)

	ticker := time.NewTicker(pollInterval)

	return Collector{
		gauges,
		counters,
		ticker,
	}
}

func (c *Collector) Start() {
	for {
		select {
		case <-c.ticker.C:
			c.Poll()
		}
	}
}

func (c *Collector) Poll() {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	c.gaugeMetrics[Alloc] = gauge(m.Alloc)
	c.gaugeMetrics[BuckHashSys] = gauge(m.BuckHashSys)
	c.gaugeMetrics[Frees] = gauge(m.Frees)
	c.gaugeMetrics[GCCPUFraction] = gauge(m.GCCPUFraction)
	c.gaugeMetrics[GCSys] = gauge(m.GCSys)
	c.gaugeMetrics[HeapAlloc] = gauge(m.HeapAlloc)
	c.gaugeMetrics[HeapIdle] = gauge(m.HeapIdle)
	c.gaugeMetrics[HeapInuse] = gauge(m.HeapInuse)
	c.gaugeMetrics[HeapObjects] = gauge(m.HeapObjects)
	c.gaugeMetrics[HeapReleased] = gauge(m.HeapReleased)
	c.gaugeMetrics[HeapSys] = gauge(m.HeapSys)
	c.gaugeMetrics[LastGC] = gauge(m.LastGC)
	c.gaugeMetrics[Lookups] = gauge(m.Lookups)
	c.gaugeMetrics[MCacheInuse] = gauge(m.MCacheInuse)
	c.gaugeMetrics[MCacheSys] = gauge(m.MCacheSys)
	c.gaugeMetrics[MSpanInuse] = gauge(m.MSpanInuse)
	c.gaugeMetrics[MSpanSys] = gauge(m.MSpanSys)
	c.gaugeMetrics[Mallocs] = gauge(m.Mallocs)
	c.gaugeMetrics[NextGC] = gauge(m.NextGC)
	c.gaugeMetrics[NumForcedGC] = gauge(m.NumForcedGC)
	c.gaugeMetrics[NumGC] = gauge(m.NumGC)
	c.gaugeMetrics[OtherSys] = gauge(m.OtherSys)
	c.gaugeMetrics[PauseTotalNs] = gauge(m.PauseTotalNs)
	c.gaugeMetrics[StackInuse] = gauge(m.StackInuse)
	c.gaugeMetrics[StackSys] = gauge(m.StackSys)
	c.gaugeMetrics[Sys] = gauge(m.Sys)
	c.gaugeMetrics[TotalAlloc] = gauge(m.TotalAlloc)
	c.gaugeMetrics[RandomValue] = gauge(rand.Intn(100))

	c.counterMetrics[PollCount]++
}
