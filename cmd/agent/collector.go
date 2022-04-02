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
	"sync"
	"time"

	"github.com/vleukhin/prom-light/internal/metrics"
)

type CollectorConfig struct {
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL " envDefault:"10s"`
	ReportTimeout  time.Duration `env:"REPORT_TIMEOUT" envDefault:"1s"`
	ServerAddr     string        `env:"ADDRESS" envDefault:"localhost:8080"`
}

type Collector struct {
	gaugeMetrics   map[string]metrics.Gauge
	counterMetrics map[string]metrics.Counter
	pollTicker     *time.Ticker
	reportTicker   *time.Ticker
	client         http.Client
	cfg            CollectorConfig
	mutex          *sync.Mutex
}

func NewCollector(config CollectorConfig) Collector {
	rand.Seed(time.Now().Unix())
	var mutex sync.Mutex

	client := http.Client{}
	client.Timeout = config.ReportTimeout

	return Collector{
		make(map[string]metrics.Gauge),
		make(map[string]metrics.Counter),
		time.NewTicker(config.PollInterval),
		time.NewTicker(config.ReportInterval),
		client,
		config,
		&mutex,
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
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Println("Polling metrics")
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	c.gaugeMetrics[metrics.Alloc] = metrics.Gauge(m.Alloc)
	c.gaugeMetrics[metrics.BuckHashSys] = metrics.Gauge(m.BuckHashSys)
	c.gaugeMetrics[metrics.Frees] = metrics.Gauge(m.Frees)
	c.gaugeMetrics[metrics.GCCPUFraction] = metrics.Gauge(m.GCCPUFraction)
	c.gaugeMetrics[metrics.GCSys] = metrics.Gauge(m.GCSys)
	c.gaugeMetrics[metrics.HeapAlloc] = metrics.Gauge(m.HeapAlloc)
	c.gaugeMetrics[metrics.HeapIdle] = metrics.Gauge(m.HeapIdle)
	c.gaugeMetrics[metrics.HeapInuse] = metrics.Gauge(m.HeapInuse)
	c.gaugeMetrics[metrics.HeapObjects] = metrics.Gauge(m.HeapObjects)
	c.gaugeMetrics[metrics.HeapReleased] = metrics.Gauge(m.HeapReleased)
	c.gaugeMetrics[metrics.HeapSys] = metrics.Gauge(m.HeapSys)
	c.gaugeMetrics[metrics.LastGC] = metrics.Gauge(m.LastGC)
	c.gaugeMetrics[metrics.Lookups] = metrics.Gauge(m.Lookups)
	c.gaugeMetrics[metrics.MCacheInuse] = metrics.Gauge(m.MCacheInuse)
	c.gaugeMetrics[metrics.MCacheSys] = metrics.Gauge(m.MCacheSys)
	c.gaugeMetrics[metrics.MSpanInuse] = metrics.Gauge(m.MSpanInuse)
	c.gaugeMetrics[metrics.MSpanSys] = metrics.Gauge(m.MSpanSys)
	c.gaugeMetrics[metrics.Mallocs] = metrics.Gauge(m.Mallocs)
	c.gaugeMetrics[metrics.NextGC] = metrics.Gauge(m.NextGC)
	c.gaugeMetrics[metrics.NumForcedGC] = metrics.Gauge(m.NumForcedGC)
	c.gaugeMetrics[metrics.NumGC] = metrics.Gauge(m.NumGC)
	c.gaugeMetrics[metrics.OtherSys] = metrics.Gauge(m.OtherSys)
	c.gaugeMetrics[metrics.PauseTotalNs] = metrics.Gauge(m.PauseTotalNs)
	c.gaugeMetrics[metrics.StackInuse] = metrics.Gauge(m.StackInuse)
	c.gaugeMetrics[metrics.StackSys] = metrics.Gauge(m.StackSys)
	c.gaugeMetrics[metrics.Sys] = metrics.Gauge(m.Sys)
	c.gaugeMetrics[metrics.TotalAlloc] = metrics.Gauge(m.TotalAlloc)
	c.gaugeMetrics[metrics.RandomValue] = metrics.Gauge(rand.Intn(100))

	c.counterMetrics[metrics.PollCount]++
}

func (c *Collector) report() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Println("Sending metrics")
	var mtrcs metrics.Metrics

	for name, v := range c.gaugeMetrics {
		value := v
		mtrcs = append(mtrcs, metrics.Metric{
			Name:  name,
			Type:  metrics.GaugeTypeName,
			Value: &value,
		})
	}
	for name, v := range c.counterMetrics {
		value := v
		mtrcs = append(mtrcs, metrics.Metric{
			Name:  name,
			Type:  metrics.CounterTypeName,
			Delta: &value,
		})
	}

	for _, m := range mtrcs {
		err := c.sendReportRequest(m)
		if err != nil {
			log.Println("Error occurred while reporting " + m.Name + " metric:" + err.Error())
			continue
		}

		if m.IsCounter() {
			c.counterMetrics[m.Name] = 0
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
