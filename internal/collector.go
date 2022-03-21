package internal

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

type gauge float64
type counter int64

const gaugeTypeName = "gauge"
const counterTypeName = "counter"

type CollectorConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ReportTimeout  time.Duration
	ServerHost     string
	ServerPort     uint32
}

type Collector struct {
	gaugeMetrics   map[string]gauge
	counterMetrics map[string]counter
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
		make(map[string]gauge),
		make(map[string]counter),
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
	fmt.Println("Poll metrics")
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

func (c *Collector) report() {
	var reportUrl string
	for name, value := range c.gaugeMetrics {
		reportUrl = c.buildMetricUrl(gaugeTypeName, name) + fmt.Sprintf("%f", value)
		err := c.sendReportRequest(reportUrl, name)
		if err != nil {
			continue
		}
	}
	for name, value := range c.counterMetrics {
		reportUrl = c.buildMetricUrl(counterTypeName, name) + fmt.Sprintf("%d", value)
		err := c.sendReportRequest(reportUrl, name)
		if err != nil {
			continue
		}
	}
}

func (c *Collector) sendReportRequest(reportUrl, metricName string) error {
	resp, err := c.client.Post(reportUrl, "text/plain", nil)
	if err != nil {
		fmt.Println("Error occurred while reporting " + metricName + " metric:" + err.Error())
		return err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Bad response while reporting " + metricName + " metric:" + strconv.Itoa(resp.StatusCode))
		return errors.New("bad response")
	}

	return nil
}

func (c *Collector) buildMetricUrl(metricType, name string) string {
	return fmt.Sprintf("http://%s:%d/update/%s/%s/", c.cfg.ServerHost, c.cfg.ServerPort, metricType, name)
}
