package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"github.com/pkg/errors"
	"hash"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/vleukhin/prom-light/internal/config"
	"github.com/vleukhin/prom-light/internal/crypt"
	"github.com/vleukhin/prom-light/internal/storage"
)

type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// App описывает сервер сбора метрик
type App struct {
	cfg    *config.ServerConfig
	str    storage.MetricsStorage
	server Server
}

// NewApp создает новый сервер сбора метрик
func NewApp(cfg *config.ServerConfig) (*App, error) {
	str, err := newStorage(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create storage")
	}

	server, err := newServer(cfg, str)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create server")
	}

	app := App{
		cfg:    cfg,
		str:    str,
		server: server,
	}

	err = str.Migrate(context.Background())
	if err != nil {
		return nil, err
	}

	return &app, nil
}

func newServer(cfg *config.ServerConfig, str storage.MetricsStorage) (Server, error) {
	var hasher hash.Hash
	var server Server
	if cfg.Key != "" {
		hasher = hmac.New(sha256.New, []byte(cfg.Key))
	}

	privateKey, err := crypt.GetPrivateKeyFromFile(cfg.CryptoKey)
	if err != nil {
		return nil, err
	}

	switch cfg.Protocol {
	case config.ProtocolHTTP:
		server = NewHTTPServer(cfg.Addr, str, hasher, privateKey, cfg.TrustedSubnet)
	case config.ProtocolGRPC:
		server = NewGRPCServer(str)
	default:
		return nil, errors.New("unknown protocol: " + cfg.Protocol)
	}

	return server, nil
}

func newStorage(cfg *config.ServerConfig) (storage.MetricsStorage, error) {
	var err error
	var str storage.MetricsStorage

	switch true {
	case cfg.DSN != "":
		str, err = storage.NewPostgresStorage(cfg.DSN, cfg.DBConnTimeout.Duration)
		if err != nil {
			return nil, err
		}
	case cfg.StoreFile != "":
		str, err = storage.NewFileStorage(cfg.StoreFile, cfg.StoreInterval.Duration, cfg.Restore)
		if err != nil {
			return nil, err
		}
	default:
		str = storage.NewMemoryStorage()
	}

	return str, err
}

// Run запукает сервер сбора метрик
func (s *App) Run(err chan<- error) {
	log.Info().Msgf("Metrics %s server listen at: %s", s.cfg.Protocol, s.cfg.Addr)
	err <- s.server.ListenAndServe()
}

// Stop останавливает сервер сбора метрик
func (s *App) Stop(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	return s.str.ShutDown(ctx)
}

func pingHandler(store storage.MetricsStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := store.Ping(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
