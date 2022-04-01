package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/vleukhin/prom-light/cmd/server/handlers"
	"github.com/vleukhin/prom-light/cmd/server/storage"
)

type ServerConfig struct {
	Addr string
	Port uint16
}

type MetricsServer struct {
	cfg ServerConfig
}

func NewMetricsServer(cfg ServerConfig) MetricsServer {
	return MetricsServer{cfg: cfg}
}

func (s MetricsServer) Run(err chan<- error) {
	addr := fmt.Sprintf("%s:%d", s.cfg.Addr, s.cfg.Port)

	fmt.Println("Metrics server listen at: " + addr)
	err <- http.ListenAndServe(addr, NewRouter(storage.NewMemoryStorage()))
}

func NewRouter(str storage.MetricsStorage) *mux.Router {
	homeHandler := handlers.NewHomeHandler(str)
	updateHandler := handlers.NewUpdateMetricHandler(str)
	updateJSONHandler := handlers.NewUpdateMetricJSONHandler(str)
	getHandler := handlers.NewGetMetricHandler(str)
	getJSONHandler := handlers.NewGetMetricJSONHandler(str)

	r := mux.NewRouter()
	r.Handle("/", homeHandler).Methods(http.MethodGet)
	r.Handle("/update", updateJSONHandler).Methods(http.MethodPost)
	r.Handle("/update/{type}/{name}/{value}", updateHandler).Methods(http.MethodPost)
	r.Handle("/value", getJSONHandler).Methods(http.MethodPost)
	r.Handle("/value/{type}/{name}", getHandler).Methods(http.MethodGet)

	return r
}
