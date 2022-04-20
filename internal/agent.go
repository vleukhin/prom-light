package internal

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

func (c *Agent) report() {
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

func (c *Agent) sendReportRequest(m metrics.Metric) error {
	c.Sign(&m)
	data, err := json.Marshal(m)
	fmt.Println(string(data))
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

func (c *Agent) Sign(m *metrics.Metric) {
	if c.hasher == nil {
		return
	}

	switch m.Type {
	case metrics.CounterTypeName:
		c.hasher.Write([]byte(fmt.Sprintf("%s:counter:%d", m.Name, m.Delta)))
	case metrics.GaugeTypeName:
		c.hasher.Write([]byte(fmt.Sprintf("%s:gauge:%d", m.Name, m.Value)))
	}

	m.Hash = hex.EncodeToString(c.hasher.Sum(nil))
	c.hasher.Reset()
}
