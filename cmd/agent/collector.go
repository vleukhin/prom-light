package main

import (
	"errors"
	"fmt"
	"github.com/vleukhin/prom-light/internal"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

type CollectorConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ReportTimeout  time.Duration
	ServerHost     string
	ServerPort     uint16
}

type Collector struct {
	gaugeMetrics   map[string]internal.Gauge
	counterMetrics map[string]internal.Counter
	pollTicker     *time.Ticker
	reportTicker   *time.Ticker
	client         http.Client
	cfg            CollectorConfig
}

func NewCollector(config CollectorConfig) Collector {
	rand.Seed(time.Now().Unix())

	client := http.Client{}
	client.Timeout = config.ReportTimeout

	return Collector{
		make(map[string]internal.Gauge),
		make(map[string]internal.Counter),
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

func (c *Collector) poll() {
	fmt.Println("Polling metrics")
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	c.gaugeMetrics[internal.Alloc] = internal.Gauge(m.Alloc)
	c.gaugeMetrics[internal.BuckHashSys] = internal.Gauge(m.BuckHashSys)
	c.gaugeMetrics[internal.Frees] = internal.Gauge(m.Frees)
	c.gaugeMetrics[internal.GCCPUFraction] = internal.Gauge(m.GCCPUFraction)
	c.gaugeMetrics[internal.GCSys] = internal.Gauge(m.GCSys)
	c.gaugeMetrics[internal.HeapAlloc] = internal.Gauge(m.HeapAlloc)
	c.gaugeMetrics[internal.HeapIdle] = internal.Gauge(m.HeapIdle)
	c.gaugeMetrics[internal.HeapInuse] = internal.Gauge(m.HeapInuse)
	c.gaugeMetrics[internal.HeapObjects] = internal.Gauge(m.HeapObjects)
	c.gaugeMetrics[internal.HeapReleased] = internal.Gauge(m.HeapReleased)
	c.gaugeMetrics[internal.HeapSys] = internal.Gauge(m.HeapSys)
	c.gaugeMetrics[internal.LastGC] = internal.Gauge(m.LastGC)
	c.gaugeMetrics[internal.Lookups] = internal.Gauge(m.Lookups)
	c.gaugeMetrics[internal.MCacheInuse] = internal.Gauge(m.MCacheInuse)
	c.gaugeMetrics[internal.MCacheSys] = internal.Gauge(m.MCacheSys)
	c.gaugeMetrics[internal.MSpanInuse] = internal.Gauge(m.MSpanInuse)
	c.gaugeMetrics[internal.MSpanSys] = internal.Gauge(m.MSpanSys)
	c.gaugeMetrics[internal.Mallocs] = internal.Gauge(m.Mallocs)
	c.gaugeMetrics[internal.NextGC] = internal.Gauge(m.NextGC)
	c.gaugeMetrics[internal.NumForcedGC] = internal.Gauge(m.NumForcedGC)
	c.gaugeMetrics[internal.NumGC] = internal.Gauge(m.NumGC)
	c.gaugeMetrics[internal.OtherSys] = internal.Gauge(m.OtherSys)
	c.gaugeMetrics[internal.PauseTotalNs] = internal.Gauge(m.PauseTotalNs)
	c.gaugeMetrics[internal.StackInuse] = internal.Gauge(m.StackInuse)
	c.gaugeMetrics[internal.StackSys] = internal.Gauge(m.StackSys)
	c.gaugeMetrics[internal.Sys] = internal.Gauge(m.Sys)
	c.gaugeMetrics[internal.TotalAlloc] = internal.Gauge(m.TotalAlloc)
	c.gaugeMetrics[internal.RandomValue] = internal.Gauge(rand.Intn(100))

	c.counterMetrics[internal.PollCount]++
}

func (c *Collector) report() {
	fmt.Println("Sending metrics")
	var reportURL string
	for name, value := range c.gaugeMetrics {
		reportURL = c.buildMetricURL(internal.GaugeTypeName.String(), name) + fmt.Sprintf("%f", value)
		err := c.sendReportRequest(reportURL, name)
		if err != nil {
			continue
		}
	}
	for name, value := range c.counterMetrics {
		reportURL = c.buildMetricURL(internal.CounterTypeName.String(), name) + fmt.Sprintf("%d", value)
		err := c.sendReportRequest(reportURL, name)
		if err != nil {
			continue
		}
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
