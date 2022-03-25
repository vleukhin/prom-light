package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/vleukhin/prom-light/cmd/server/handlers"
	"github.com/vleukhin/prom-light/cmd/server/storage"
	"net/http"
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

func NewRouter(str handlers.MetricsStorage) *mux.Router {
	updateHandler := handlers.NewUpdateMetricHandler(str)
	getHandler := handlers.NewGetMetricHandler(str)

	r := mux.NewRouter()
	r.Handle("/update/{type}/{name}/{value}", updateHandler).Methods(http.MethodPost)
	r.Handle("/value/{type}/{name}", getHandler).Methods(http.MethodGet)

	return r
}
