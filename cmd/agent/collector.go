package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/vleukhin/prom-light/internal/metrics"
)

type CollectorConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ReportTimeout  time.Duration
	ServerHost     string
	ServerPort     uint16
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

	fmt.Println("Polling metrics")
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

	fmt.Println("Sending metrics")
	var reportURL string
	for name, value := range c.gaugeMetrics {
		reportURL = c.buildMetricURL(metrics.GaugeTypeName.String(), name) + fmt.Sprintf("%f", value)
		err := c.sendReportRequest(reportURL, name)
		if err != nil {
			continue
		}
	}
	for name, value := range c.counterMetrics {
		reportURL = c.buildMetricURL(metrics.CounterTypeName.String(), name) + fmt.Sprintf("%d", value)
		err := c.sendReportRequest(reportURL, name)
		if err != nil {
			continue
		}
		c.counterMetrics[name] = 0
	}
}

func (c *Collector) sendReportRequest(reportURL, metricName string) error {
	resp, err := c.client.Post(reportURL, "text/plain", nil)
	if err != nil {
		fmt.Println("Error occurred while reporting " + metricName + " metric:" + err.Error())
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		fmt.Println("Error while closing response body:" + err.Error())
		return err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Bad response while reporting " + metricName + " metric:" + strconv.Itoa(resp.StatusCode))
		return errors.New("bad response")
	}

	return nil
}

func (c *Collector) buildMetricURL(metricType, name string) string {
	return fmt.Sprintf("http://%s:%d/update/%s/%s/", c.cfg.ServerHost, c.cfg.ServerPort, metricType, name)
}
