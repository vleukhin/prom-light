package internal

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/vleukhin/prom-light/internal/metrics"
	"github.com/vleukhin/prom-light/internal/storage"
)

type Agent struct {
	storage      storage.MetricsStorage
	pollTicker   *time.Ticker
	reportTicker *time.Ticker
	client       http.Client
	cfg          *AgentConfig
	hasher       hash.Hash
}

func NewAgent(config *AgentConfig) Agent {
	rand.Seed(time.Now().Unix())

	client := http.Client{}
	client.Timeout = config.ReportTimeout

	agent := Agent{
		storage.NewMemoryStorage(),
		time.NewTicker(config.PollInterval),
		time.NewTicker(config.ReportInterval),
		client,
		config,
		nil,
	}

	if config.Key != "" {
		agent.hasher = hmac.New(sha256.New, []byte(config.Key))
	}

	return agent
}

func (c *Agent) Start() {
	for {
		select {
		case <-c.pollTicker.C:
			c.poll()
		case <-c.reportTicker.C:
			c.report()
		}
	}
}

func (c *Agent) Stop() {
	c.pollTicker.Stop()
	c.reportTicker.Stop()
}

func (c *Agent) poll() {
	log.Println("Polling metrics")
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	ctx := context.TODO()

	c.storage.SetGauge(ctx, metrics.Alloc, metrics.Gauge(m.Alloc))
	c.storage.SetGauge(ctx, metrics.BuckHashSys, metrics.Gauge(m.BuckHashSys))
	c.storage.SetGauge(ctx, metrics.Frees, metrics.Gauge(m.Frees))
	c.storage.SetGauge(ctx, metrics.GCCPUFraction, metrics.Gauge(m.GCCPUFraction))
	c.storage.SetGauge(ctx, metrics.GCSys, metrics.Gauge(m.GCSys))
	c.storage.SetGauge(ctx, metrics.HeapAlloc, metrics.Gauge(m.HeapAlloc))
	c.storage.SetGauge(ctx, metrics.HeapIdle, metrics.Gauge(m.HeapIdle))
	c.storage.SetGauge(ctx, metrics.HeapInuse, metrics.Gauge(m.HeapInuse))
	c.storage.SetGauge(ctx, metrics.HeapObjects, metrics.Gauge(m.HeapObjects))
	c.storage.SetGauge(ctx, metrics.HeapReleased, metrics.Gauge(m.HeapReleased))
	c.storage.SetGauge(ctx, metrics.HeapSys, metrics.Gauge(m.HeapSys))
	c.storage.SetGauge(ctx, metrics.LastGC, metrics.Gauge(m.LastGC))
	c.storage.SetGauge(ctx, metrics.Lookups, metrics.Gauge(m.Lookups))
	c.storage.SetGauge(ctx, metrics.MCacheInuse, metrics.Gauge(m.MCacheInuse))
	c.storage.SetGauge(ctx, metrics.MCacheSys, metrics.Gauge(m.MCacheSys))
	c.storage.SetGauge(ctx, metrics.MSpanInuse, metrics.Gauge(m.MSpanInuse))
	c.storage.SetGauge(ctx, metrics.MSpanSys, metrics.Gauge(m.MSpanSys))
	c.storage.SetGauge(ctx, metrics.Mallocs, metrics.Gauge(m.Mallocs))
	c.storage.SetGauge(ctx, metrics.NextGC, metrics.Gauge(m.NextGC))
	c.storage.SetGauge(ctx, metrics.NumForcedGC, metrics.Gauge(m.NumForcedGC))
	c.storage.SetGauge(ctx, metrics.NumGC, metrics.Gauge(m.NumGC))
	c.storage.SetGauge(ctx, metrics.OtherSys, metrics.Gauge(m.OtherSys))
	c.storage.SetGauge(ctx, metrics.PauseTotalNs, metrics.Gauge(m.PauseTotalNs))
	c.storage.SetGauge(ctx, metrics.StackInuse, metrics.Gauge(m.StackInuse))
	c.storage.SetGauge(ctx, metrics.StackSys, metrics.Gauge(m.StackSys))
	c.storage.SetGauge(ctx, metrics.Sys, metrics.Gauge(m.Sys))
	c.storage.SetGauge(ctx, metrics.TotalAlloc, metrics.Gauge(m.TotalAlloc))
	c.storage.SetGauge(ctx, metrics.RandomValue, metrics.Gauge(rand.Intn(100)))
	c.storage.SetGauge(ctx, metrics.StaticGauge, 100)

	c.storage.IncCounter(ctx, metrics.PollCount, 1)
}

func (c *Agent) report() {
	ctx := context.TODO()
	log.Println("Sending metrics")
	mtrcs := c.storage.GetAllMetrics(ctx, true)

	for _, m := range mtrcs {
		err := c.sendReportRequest(m)
		if err != nil {
			log.Println("Error occurred while reporting " + m.Name + " metric:" + err.Error())
			if m.IsCounter() {
				c.storage.IncCounter(ctx, m.Name, *m.Delta)
			}
			continue
		}
	}
}

func (c *Agent) sendReportRequest(m metrics.Metric) error {
	m.Sign(c.hasher)

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
