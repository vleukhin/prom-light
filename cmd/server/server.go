package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/vleukhin/prom-light/cmd/server/handlers"
	"github.com/vleukhin/prom-light/cmd/server/storage"
)

type ServerConfig struct {
	Addr          string        `env:"ADDRESS" envDefault:"localhost:8080"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
}

type MetricsServer struct {
	cfg ServerConfig
	str storage.MetricsStorage
}

func NewMetricsServer(cfg ServerConfig) (MetricsServer, error) {
	var err error
	var str storage.MetricsStorage

	if cfg.StoreFile == "" {
		str = storage.NewMemoryStorage()
	} else {
		str, err = storage.NewFileStorage(cfg.StoreFile, cfg.StoreInterval, cfg.Restore)
		if err != nil {
			return MetricsServer{}, err
		}
	}
	return MetricsServer{
		cfg: cfg,
		str: str,
	}, nil
}

func (s MetricsServer) Run(err chan<- error) {
	log.Println("Metrics server listen at: " + s.cfg.Addr)
	err <- http.ListenAndServe(s.cfg.Addr, NewRouter(s.str))
}

func (s MetricsServer) Stop() {
	s.str.ShutDown()
}

func NewRouter(str storage.MetricsStorage) *mux.Router {
	homeHandler := handlers.NewHomeHandler(str)
	updateHandler := handlers.NewUpdateMetricHandler(str)
	updateJSONHandler := handlers.NewUpdateMetricJSONHandler(str)
	getHandler := handlers.NewGetMetricHandler(str)
	getJSONHandler := handlers.NewGetMetricJSONHandler(str)

	r := mux.NewRouter()
	r.Handle("/", homeHandler).Methods(http.MethodGet)
	r.Handle("/update/", updateJSONHandler).Methods(http.MethodPost)
	r.Handle("/update/{type}/{name}/{value}", updateHandler).Methods(http.MethodPost)
	r.Handle("/value/", getJSONHandler).Methods(http.MethodPost)
	r.Handle("/value/{type}/{name}", getHandler).Methods(http.MethodGet)

	return r
}
