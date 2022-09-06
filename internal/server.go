package internal

import (
	"context"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"hash"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/vleukhin/prom-light/internal/crypt"
	"github.com/vleukhin/prom-light/internal/middlewares"

	"github.com/vleukhin/prom-light/internal/handlers"
	"github.com/vleukhin/prom-light/internal/storage"
)

// MetricsServer описывает сервер сбора метрик
type MetricsServer struct {
	cfg        *ServerConfig
	str        storage.MetricsStorage
	hasher     hash.Hash
	PrivateKey *rsa.PrivateKey
}

// NewMetricsServer создает новый сервер сбора метрик
func NewMetricsServer(config *ServerConfig) (*MetricsServer, error) {
	var err error
	var str storage.MetricsStorage

	switch true {
	case config.DSN != "":
		str, err = storage.NewPostgresStorage(config.DSN, config.DBConnTimeout)
		if err != nil {
			return nil, err
		}
	case config.StoreFile != "":
		str, err = storage.NewFileStorage(config.StoreFile, config.StoreInterval, config.Restore)
		if err != nil {
			return nil, err
		}
	default:
		str = storage.NewMemoryStorage()
	}

	server := MetricsServer{
		config,
		str,
		nil,
		nil,
	}

	if err := server.setPrivateKey(); err != nil {
		return nil, err
	}

	err = server.migrate()
	if err != nil {
		return nil, err
	}

	if config.Key != "" {
		server.hasher = hmac.New(sha256.New, []byte(config.Key))
	}

	return &server, nil
}

func (s *MetricsServer) setPrivateKey() error {
	if s.cfg.CryptoKey == "" {
		return nil
	}
	b, err := os.ReadFile(s.cfg.CryptoKey)
	if err != nil {
		return err
	}
	s.PrivateKey, err = crypt.BytesToPrivateKey(b)
	return err
}

// Run запукает сервер сбора метрик
func (s *MetricsServer) Run(err chan<- error) {
	log.Info().Msg("Metrics server listen at: " + s.cfg.Addr)
	err <- http.ListenAndServe(s.cfg.Addr, NewRouter(s.str, s.hasher, s.PrivateKey))
}

// Stop останавливает сервер сбора метрик
func (s *MetricsServer) Stop() error {
	return s.str.ShutDown(context.TODO())
}

// NewRouter создает новый роутер
func NewRouter(str storage.MetricsStorage, hasher hash.Hash, key *rsa.PrivateKey) *mux.Router {
	homeHandler := handlers.NewHomeHandler(str)
	updateHandler := handlers.NewUpdateMetricHandler(str)
	updateJSONHandler := handlers.NewUpdateMetricJSONHandler(str, hasher)
	updateBatchHandler := handlers.NewUpdateMetricsBatchHandler(str, hasher)
	getHandler := handlers.NewGetMetricHandler(str)
	getJSONHandler := handlers.NewGetMetricJSONHandler(str, hasher)

	r := mux.NewRouter()
	r.Use(middlewares.GZIPEncode)
	r.Use(middlewares.NewDecryptMiddleware(key).Handle)
	r.Handle("/", homeHandler).Methods(http.MethodGet, http.MethodHead)
	r.Handle("/update/", updateJSONHandler).Methods(http.MethodPost)
	r.Handle("/updates/", updateBatchHandler).Methods(http.MethodPost)
	r.Handle("/update/{type}/{name}/{value}", updateHandler).Methods(http.MethodPost)
	r.Handle("/value/", getJSONHandler).Methods(http.MethodPost)
	r.Handle("/value/{type}/{name}", getHandler).Methods(http.MethodGet, http.MethodHead)
	r.Handle("/ping", pingHandler(str)).Methods(http.MethodGet, http.MethodHead)

	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	return r
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

func (s *MetricsServer) migrate() error {
	return s.str.Migrate(context.Background())
}
