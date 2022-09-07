package internal

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	mrand "math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/vleukhin/prom-light/internal/config"

	"github.com/vleukhin/prom-light/internal/crypt"

	"github.com/rs/zerolog/log"

	"github.com/vleukhin/prom-light/internal/metrics"
	"github.com/vleukhin/prom-light/internal/pollers"
	"github.com/vleukhin/prom-light/internal/storage"
)

type Poller interface {
	// Poll сбор метрик
	Poll() (metrics.Metrics, error)
}

// Agent описывает агент для сбра метрик
type Agent struct {
	storage      storage.MetricsStorage
	reportTicker *time.Ticker
	pollTicker   *time.Ticker
	client       http.Client
	cfg          *config.AgentConfig
	pollers      []Poller
	hasher       hash.Hash
	cancel       context.CancelFunc
	publicKey    *rsa.PublicKey
}

// NewAgent создаёт новый агент для сбора метрик
func NewAgent(config *config.AgentConfig) (*Agent, error) {
	mrand.Seed(time.Now().Unix())

	client := http.Client{}
	client.Timeout = config.ReportTimeout.Duration

	agent := Agent{
		storage:      storage.NewMemoryStorage(),
		reportTicker: time.NewTicker(config.ReportInterval.Duration),
		pollTicker:   time.NewTicker(config.PollInterval.Duration),
		client:       client,
		cfg:          config,
	}

	if err := agent.setPublicKey(); err != nil {
		return nil, err
	}

	if config.Key != "" {
		agent.hasher = hmac.New(sha256.New, []byte(config.Key))
	}

	agent.pollers = append(agent.pollers, pollers.MemStatsPoller{})
	agent.pollers = append(agent.pollers, pollers.PsPoller{})

	return &agent, nil
}

func (c *Agent) setPublicKey() error {
	if c.cfg.CryptoKey == "" {
		return nil
	}
	b, err := os.ReadFile(c.cfg.CryptoKey)
	if err != nil {
		return err
	}
	c.publicKey, err = crypt.BytesToPublicKey(b)
	return err
}

// Start запускает сбор и отправку метрик
func (c *Agent) Start(ctx context.Context, cancel context.CancelFunc) {
	c.cancel = cancel
	metricsCh := make(chan metrics.Metrics)

	go c.poll(ctx, metricsCh)
	go c.storeMetrics(ctx, metricsCh)

reportLoop:
	for range c.reportTicker.C {
		select {
		case <-ctx.Done():
			break reportLoop
		default:
			c.report(ctx)
		}
	}

	c.Stop()
}

func (c *Agent) poll(ctx context.Context, metricsCh chan<- metrics.Metrics) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().Msgf("poll() panics: %v", r)
			c.Stop()
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.pollTicker.C:
			for _, p := range c.pollers {
				mtrcs, err := p.Poll()
				if err != nil {
					log.Error().Err(err).Msgf("Failed to poll metrics from poller")
					continue
				}
				metricsCh <- mtrcs
			}
		}
	}
}

func (c *Agent) storeMetrics(ctx context.Context, metricsCh chan metrics.Metrics) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().Msgf("storeMetrics() panics: %v", r)
			c.Stop()
		}
	}()
	for m := range metricsCh {
		err := c.storage.SetMetrics(ctx, m)
		if err != nil {
			log.Error().Err(err).Msg("Failed to store metrics")
		}
	}
}

// Stop останавливает сбор и отправку метрик
func (c *Agent) Stop() {
	log.Info().Msg("Stopping agent")
	c.reportTicker.Stop()
	c.cancel()
}

// report отправляет собранные метрики на сервер
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

// sendReportRequest отправляет запрос на сервер метрик
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

// sendReportRequest отправляет batch запрос на сервер метрик
func (c *Agent) sendReportBatchRequest(m metrics.Metrics) error {
	data, err := c.encrypt(m)
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

// encrypt encrypts metrics with public key
func (c *Agent) encrypt(m metrics.Metrics) ([]byte, error) {
	data, err := json.Marshal(m.Sign(c.hasher))
	if err != nil {
		return nil, err
	}

	if c.publicKey == nil {
		return data, nil
	}

	return crypt.EncryptOAEP(c.publicKey, data, nil)
}
