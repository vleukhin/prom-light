package main

import (
	"fmt"
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

func newMetricsServer(cfg ServerConfig) MetricsServer {
	return MetricsServer{cfg: cfg}
}

func (s MetricsServer) run(err chan<- error) {
	addr := fmt.Sprintf("%s:%d", s.cfg.Addr, s.cfg.Port)
	h := handlers.NewUpdateMetricHandler(storage.NewMemoryStorage())
	http.Handle("/update/", h)
	fmt.Println("Metrics server listen at: " + addr)
	err <- http.ListenAndServe(addr, nil)
}
