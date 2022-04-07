package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/vleukhin/prom-light/cmd/server/storage"
	"github.com/vleukhin/prom-light/internal/metrics"
)

type CollectorConfig struct {
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	ReportTimeout  time.Duration `env:"REPORT_TIMEOUT"`
	ServerAddr     string        `env:"ADDRESS"`
}

type Collector struct {
	storage      storage.MetricsStorage
	pollTicker   *time.Ticker
	reportTicker *time.Ticker
	client       http.Client
	cfg          CollectorConfig
}

func NewCollector(config CollectorConfig) Collector {
	rand.Seed(time.Now().Unix())

	client := http.Client{}
	client.Timeout = config.ReportTimeout

	return Collector{
		storage.NewMemoryStorage(),
		time.NewTicker(config.PollInterval),
		time.NewTicker(config.ReportInterval),
		client,
		config,
	}
}

func (c *Collector) Start() {
	for {
		select {
		case <-c.pollTicker.C:
			c.poll()
		case <-c.reportTicker.C:
			c.report()
		}
	}
}

func (c *Collector) Stop() {
	c.pollTicker.Stop()
	c.reportTicker.Stop()
}

func (c *Collector) poll() {
	log.Println("Polling metrics")
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	c.storage.SetGauge(metrics.Alloc, metrics.Gauge(m.Alloc))
	c.storage.SetGauge(metrics.BuckHashSys, metrics.Gauge(m.BuckHashSys))
	c.storage.SetGauge(metrics.Frees, metrics.Gauge(m.Frees))
	c.storage.SetGauge(metrics.GCCPUFraction, metrics.Gauge(m.GCCPUFraction))
	c.storage.SetGauge(metrics.GCSys, metrics.Gauge(m.GCSys))
	c.storage.SetGauge(metrics.HeapAlloc, metrics.Gauge(m.HeapAlloc))
	c.storage.SetGauge(metrics.HeapIdle, metrics.Gauge(m.HeapIdle))
	c.storage.SetGauge(metrics.HeapInuse, metrics.Gauge(m.HeapInuse))
	c.storage.SetGauge(metrics.HeapObjects, metrics.Gauge(m.HeapObjects))
	c.storage.SetGauge(metrics.HeapReleased, metrics.Gauge(m.HeapReleased))
	c.storage.SetGauge(metrics.HeapSys, metrics.Gauge(m.HeapSys))
	c.storage.SetGauge(metrics.LastGC, metrics.Gauge(m.LastGC))
	c.storage.SetGauge(metrics.Lookups, metrics.Gauge(m.Lookups))
	c.storage.SetGauge(metrics.MCacheInuse, metrics.Gauge(m.MCacheInuse))
	c.storage.SetGauge(metrics.MCacheSys, metrics.Gauge(m.MCacheSys))
	c.storage.SetGauge(metrics.MSpanInuse, metrics.Gauge(m.MSpanInuse))
	c.storage.SetGauge(metrics.MSpanSys, metrics.Gauge(m.MSpanSys))
	c.storage.SetGauge(metrics.Mallocs, metrics.Gauge(m.Mallocs))
	c.storage.SetGauge(metrics.NextGC, metrics.Gauge(m.NextGC))
	c.storage.SetGauge(metrics.NumForcedGC, metrics.Gauge(m.NumForcedGC))
	c.storage.SetGauge(metrics.NumGC, metrics.Gauge(m.NumGC))
	c.storage.SetGauge(metrics.OtherSys, metrics.Gauge(m.OtherSys))
	c.storage.SetGauge(metrics.PauseTotalNs, metrics.Gauge(m.PauseTotalNs))
	c.storage.SetGauge(metrics.StackInuse, metrics.Gauge(m.StackInuse))
	c.storage.SetGauge(metrics.StackSys, metrics.Gauge(m.StackSys))
	c.storage.SetGauge(metrics.Sys, metrics.Gauge(m.Sys))
	c.storage.SetGauge(metrics.TotalAlloc, metrics.Gauge(m.TotalAlloc))
	c.storage.SetGauge(metrics.RandomValue, metrics.Gauge(rand.Intn(100)))

	c.storage.IncCounter(metrics.PollCount, 1)
}

func (c *Collector) report() {
	log.Println("Sending metrics")
	mtrcs := c.storage.GetAllMetrics(true)

	for _, m := range mtrcs {
		err := c.sendReportRequest(m)
		if err != nil {
			log.Println("Error occurred while reporting " + m.Name + " metric:" + err.Error())
			if m.IsCounter() {
				c.storage.IncCounter(m.Name, *m.Delta)
			}
			continue
		}
	}
}

func (c *Collector) sendReportRequest(m metrics.Metric) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(fmt.Sprintf("http://%s/update/", c.cfg.ServerAddr), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("bad response while reporting: " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}
