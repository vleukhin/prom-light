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
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/vleukhin/prom-light/internal/metrics"
	"github.com/vleukhin/prom-light/internal/pollers"
	"github.com/vleukhin/prom-light/internal/storage"
)

type Poller interface {
	Poll() metrics.Metrics
}

type Agent struct {
	storage      storage.MetricsStorage
	reportTicker *time.Ticker
	pollTicker   *time.Ticker
	client       http.Client
	cfg          *AgentConfig
	pollers      []Poller
	hasher       hash.Hash
}

func NewAgent(config *AgentConfig) Agent {
	rand.Seed(time.Now().Unix())

	client := http.Client{}
	client.Timeout = config.ReportTimeout

	agent := Agent{
		storage:      storage.NewMemoryStorage(),
		reportTicker: time.NewTicker(config.ReportInterval),
		pollTicker:   time.NewTicker(config.PollInterval),
		client:       client,
		cfg:          config,
	}

	if config.Key != "" {
		agent.hasher = hmac.New(sha256.New, []byte(config.Key))
	}

	agent.pollers = append(agent.pollers, pollers.MemStatsPoller{})
	agent.pollers = append(agent.pollers, pollers.PsPoller{})

	return agent
}

func (c *Agent) Start(ctx context.Context) {
	metricsCh := make(chan metrics.Metrics)

	go c.poll(ctx, metricsCh)
	go c.storeMetrics(ctx, metricsCh)

	for range c.reportTicker.C {
		c.report(ctx)
	}
}

func (c *Agent) poll(ctx context.Context, metricsCh chan<- metrics.Metrics) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.pollTicker.C:
			for _, p := range c.pollers {
				metricsCh <- p.Poll()
			}
		}
	}
}

func (c *Agent) storeMetrics(ctx context.Context, metricsCh chan metrics.Metrics) {
	for m := range metricsCh {
		err := c.storage.SetMetrics(ctx, m)
		if err != nil {
			log.Error().Err(err).Msg("Failed to store metrics")
		}
	}
}

func (c *Agent) Stop() {
	c.reportTicker.Stop()
}

func (c *Agent) report(ctx context.Context) {
	mtrcs, err := c.storage.GetAllMetrics(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get metrics to report")
	}
	log.Info().Msg("Sending metrics")
	if c.cfg.BatchMode {
		err := c.sendReportBatchRequest(mtrcs)
		if err != nil {
			log.Error().Msg("Error occurred while reporting batch of metrics:" + err.Error())
		}
	} else {
		for _, m := range mtrcs {
			err := c.sendReportRequest(m)
			if err != nil {
				log.Error().Msg("Error occurred while reporting " + m.Name + " metric:" + err.Error())
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
