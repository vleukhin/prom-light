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
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

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
	gauges := make(map[string]metrics.Gauge)
	counters := make(map[string]metrics.Counter)
	m := &runtime.MemStats{}
	ctx := context.TODO()

	log.Info().Msg("Polling metrics")
	runtime.ReadMemStats(m)

	gauges[metrics.Alloc] = metrics.Gauge(m.Alloc)
	gauges[metrics.BuckHashSys] = metrics.Gauge(m.BuckHashSys)
	gauges[metrics.Frees] = metrics.Gauge(m.Frees)
	gauges[metrics.GCCPUFraction] = metrics.Gauge(m.GCCPUFraction)
	gauges[metrics.GCSys] = metrics.Gauge(m.GCSys)
	gauges[metrics.HeapAlloc] = metrics.Gauge(m.HeapAlloc)
	gauges[metrics.HeapIdle] = metrics.Gauge(m.HeapIdle)
	gauges[metrics.HeapInuse] = metrics.Gauge(m.HeapInuse)
	gauges[metrics.HeapObjects] = metrics.Gauge(m.HeapObjects)
	gauges[metrics.HeapReleased] = metrics.Gauge(m.HeapReleased)
	gauges[metrics.HeapSys] = metrics.Gauge(m.HeapSys)
	gauges[metrics.LastGC] = metrics.Gauge(m.LastGC)
	gauges[metrics.Lookups] = metrics.Gauge(m.Lookups)
	gauges[metrics.MCacheInuse] = metrics.Gauge(m.MCacheInuse)
	gauges[metrics.MCacheSys] = metrics.Gauge(m.MCacheSys)
	gauges[metrics.MSpanInuse] = metrics.Gauge(m.MSpanInuse)
	gauges[metrics.MSpanSys] = metrics.Gauge(m.MSpanSys)
	gauges[metrics.Mallocs] = metrics.Gauge(m.Mallocs)
	gauges[metrics.NextGC] = metrics.Gauge(m.NextGC)
	gauges[metrics.NumForcedGC] = metrics.Gauge(m.NumForcedGC)
	gauges[metrics.NumGC] = metrics.Gauge(m.NumGC)
	gauges[metrics.OtherSys] = metrics.Gauge(m.OtherSys)
	gauges[metrics.PauseTotalNs] = metrics.Gauge(m.PauseTotalNs)
	gauges[metrics.StackInuse] = metrics.Gauge(m.StackInuse)
	gauges[metrics.StackSys] = metrics.Gauge(m.StackSys)
	gauges[metrics.Sys] = metrics.Gauge(m.Sys)
	gauges[metrics.TotalAlloc] = metrics.Gauge(m.TotalAlloc)
	gauges[metrics.RandomValue] = metrics.Gauge(rand.Intn(100))

	counters[metrics.PollCount] = 1

	mtrcs := make(metrics.Metrics, len(gauges)+len(counters))
	var i int

	for name, v := range gauges {
		value := v
		mtrcs[i] = metrics.Metric{
			Name:  name,
			Type:  metrics.GaugeTypeName,
			Value: &value,
		}
		i++
	}

	for name, v := range counters {
		value := v
		mtrcs[i] = metrics.Metric{
			Name:  name,
			Type:  metrics.CounterTypeName,
			Delta: &value,
		}
		i++
	}

	if err := c.storage.SetMetrics(ctx, mtrcs); err != nil {
		log.Error().Msg("Failed to set metrics in storage")
	}
}

func (c *Agent) report() {
	ctx := context.TODO()
	log.Info().Msg("Sending metrics")
	mtrcs, err := c.storage.GetAllMetrics(ctx, true)
	if err != nil {
		log.Error().Msg("Failed to get metrics")
		return
	}

	if c.cfg.BatchMode {
		err := c.sendReportBatchRequest(mtrcs)
		if err != nil {
			log.Error().Msg("Error occurred while reporting batch of metrics:" + err.Error())
			for _, m := range mtrcs {
				if m.IsCounter() {
					if err := c.storage.SetMetric(ctx, m); err != nil {
						log.Error().Msg(fmt.Sprintf("Failed to inc counter %s: %s", m.Name, err.Error()))
					}
				}
			}
		}
	} else {
		for _, m := range mtrcs {
			err := c.sendReportRequest(m)
			if err != nil {
				log.Error().Msg("Error occurred while reporting " + m.Name + " metric:" + err.Error())
				if m.IsCounter() {
					if err := c.storage.SetMetric(ctx, m); err != nil {
						log.Error().Msg(fmt.Sprintf("Failed to inc counter %s: %s", m.Name, err.Error()))
					}
				}
				continue
			}
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

func (c *Agent) sendReportBatchRequest(m metrics.Metrics) error {
	data, err := json.Marshal(m.Sign(c.hasher))
	if err != nil {
		return err
	}

	resp, err := c.client.Post(fmt.Sprintf("http://%s/updates/", c.cfg.ServerAddr), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("bad response while batch reporting: " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}
