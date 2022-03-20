package internal

import (
	"fmt"
	"time"
)

type gauge float64
type counter int64

type Collector struct {
	gaugeMetrics   map[string]gauge
	counterMetrics map[string]counter
	ticker         *time.Ticker
}

func getGaugeMetricsNames() []string {
	return []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
		"RandomValue",
	}
}
func getCounterMetricsNames() []string {
	return []string{
		"PollCount",
	}
}

func NewCollector(pollInterval, _ time.Duration) Collector {
	gauges := make(map[string]gauge)
	counters := make(map[string]counter)

	for _, name := range getGaugeMetricsNames() {
		gauges[name] = 0
	}
	for _, name := range getCounterMetricsNames() {
		counters[name] = 0
	}

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
	fmt.Println("collect metrics")
}
