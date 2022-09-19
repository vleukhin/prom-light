package agent

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"hash"
	mrand "math/rand"
	"net"
	"time"

	"github.com/pkg/errors"

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

type Client interface {
	SendMetricToServer(ctx context.Context, m metrics.Metric) error
	SendBatchMetricsToServer(ctx context.Context, m metrics.Metrics) error
	ShutDown() error
}

// App описывает агент для сбра метрик
type App struct {
	storage      storage.MetricsStorage
	reportTicker *time.Ticker
	pollTicker   *time.Ticker
	client       Client
	cfg          *config.AgentConfig
	pollers      []Poller
	hasher       hash.Hash
	cancel       context.CancelFunc
}

// NewApp создаёт новый агент для сбора метрик
func NewApp(config *config.AgentConfig) (*App, error) {
	mrand.Seed(time.Now().Unix())

	client, err := newClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create agent client")
	}

	agent := App{
		storage:      storage.NewMemoryStorage(),
		reportTicker: time.NewTicker(config.ReportInterval.Duration),
		pollTicker:   time.NewTicker(config.PollInterval.Duration),
		client:       client,
		cfg:          config,
	}

	if config.Key != "" {
		agent.hasher = hmac.New(sha256.New, []byte(config.Key))
	}

	agent.pollers = append(agent.pollers, pollers.MemStatsPoller{})
	agent.pollers = append(agent.pollers, pollers.PsPoller{})

	return &agent, nil
}

func newClient(cfg *config.AgentConfig) (Client, error) {
	var client Client
	addr, err := detectIP()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to detect host IP")
	}

	key, err := crypt.GetPublicKeyFromFile(cfg.CryptoKey)
	if err != nil {
		return nil, err
	}

	switch cfg.Protocol {
	case config.ProtocolHTTP:
		client = NewHTTPClient(cfg.ServerAddr, addr.IP, cfg.ReportTimeout.Duration, key)
	case config.ProtocolGRPC:
		client, err = NewGRPCClient(cfg.ServerAddr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create GRPC client")
		}
	default:
		return nil, errors.New("unknown protocol: " + cfg.Protocol)
	}

	return client, nil
}

// Start запускает сбор и отправку метрик
func (c *App) Start(ctx context.Context, cancel context.CancelFunc) {
	log.Info().Msgf("%s agent started", c.cfg.Protocol)
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

	c.Stop(ctx)
}

func (c *App) poll(ctx context.Context, metricsCh chan<- metrics.Metrics) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().Msgf("poll() panics: %v", r)
			c.Stop(ctx)
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

func (c *App) storeMetrics(ctx context.Context, metricsCh chan metrics.Metrics) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().Msgf("storeMetrics() panics: %v", r)
			c.Stop(ctx)
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
func (c *App) Stop(ctx context.Context) {
	log.Info().Msg("Stopping agent")
	c.report(ctx)
	c.reportTicker.Stop()
	err := c.client.ShutDown()
	if err != nil {
		log.Error().Err(err).Msg("got error while stopping client")
	}
	c.cancel()
}

// report отправляет собранные метрики на сервер
func (c *App) report(ctx context.Context) {
	mtrcs, err := c.storage.GetAllMetrics(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get metrics to report")
	}
	log.Info().Msg("Sending metrics")
	mtrcs = mtrcs.Sign(c.hasher)
	if c.cfg.BatchMode {
		err := c.client.SendBatchMetricsToServer(ctx, mtrcs)
		if err != nil {
			log.Error().Msg("Error occurred while reporting batch of metrics:" + err.Error())
		}
	} else {
		for _, m := range mtrcs {
			err := c.client.SendMetricToServer(ctx, m)
			if err != nil {
				log.Error().Msg("Error occurred while reporting " + m.Name + " metric:" + err.Error())
			}
		}
	}
}

func detectIP() (*net.UDPAddr, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	// handle err...
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr), nil
}
