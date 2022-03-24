package main

import (
	"fmt"
	"net/http"
)

type ServerConfig struct {
	Addr string
	Port uint16
}

type MetrcisServer struct {
	cfg ServerConfig
}

func newMetricsServer(cfg ServerConfig) MetrcisServer {
	return MetrcisServer{cfg: cfg}
}

func (s MetrcisServer) run(err chan<- error) {
	addr := fmt.Sprintf("%s:%d", s.cfg.Addr, s.cfg.Port)
	fmt.Println("Metrics server listen at: " + addr)
	err <- http.ListenAndServe(addr, nil)
}
