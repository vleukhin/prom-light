package pollers

import (
	"github.com/vleukhin/prom-light/internal/metrics"
	"math/rand"
	"runtime"
)

type MemStatsPoller struct {
}

func (p MemStatsPoller) Poll() metrics.Metrics {
	m := &runtime.MemStats{}
	mtrcs := make(metrics.Metrics, 0, 29)
	runtime.ReadMemStats(m)

	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("Alloc", metrics.Gauge(m.Alloc)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("BuckHashSys", metrics.Gauge(m.BuckHashSys)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("Frees", metrics.Gauge(m.Frees)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("GCCPUFraction", metrics.Gauge(m.GCCPUFraction)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("GCSys", metrics.Gauge(m.GCSys)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("HeapAlloc", metrics.Gauge(m.HeapAlloc)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("HeapIdle", metrics.Gauge(m.HeapIdle)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("HeapInuse", metrics.Gauge(m.HeapInuse)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("HeapObjects", metrics.Gauge(m.HeapObjects)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("HeapReleased", metrics.Gauge(m.HeapReleased)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("HeapSys", metrics.Gauge(m.HeapSys)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("LastGC", metrics.Gauge(m.LastGC)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("Lookups", metrics.Gauge(m.Lookups)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("MCacheInuse", metrics.Gauge(m.MCacheInuse)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("MCacheSys", metrics.Gauge(m.MCacheSys)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("MSpanInuse", metrics.Gauge(m.MSpanInuse)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("MSpanSys", metrics.Gauge(m.MSpanSys)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("Mallocs", metrics.Gauge(m.Mallocs)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("NextGC", metrics.Gauge(m.NextGC)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("NumForcedGC", metrics.Gauge(m.NumForcedGC)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("NumGC", metrics.Gauge(m.NumGC)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("OtherSys", metrics.Gauge(m.OtherSys)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("PauseTotalNs", metrics.Gauge(m.PauseTotalNs)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("StackInuse", metrics.Gauge(m.StackInuse)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("StackSys", metrics.Gauge(m.StackSys)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("Sys", metrics.Gauge(m.Sys)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("TotalAlloc", metrics.Gauge(m.TotalAlloc)))
	mtrcs = append(mtrcs, metrics.MakeGaugeMetric("RandomValue", metrics.Gauge(rand.Intn(100))))

	mtrcs = append(mtrcs, metrics.MakeCounterMetric("PollCount", 1))

	return mtrcs
}
