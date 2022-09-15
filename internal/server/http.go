package server

import (
	"crypto/rsa"
	"github.com/gorilla/mux"
	"github.com/vleukhin/prom-light/internal/http-handlers"
	"github.com/vleukhin/prom-light/internal/middlewares"
	"github.com/vleukhin/prom-light/internal/storage"
	"hash"
	"net"
	"net/http"
)

func NewHTTPServer(
	addr string,
	str storage.MetricsStorage,
	hasher hash.Hash,
	key *rsa.PrivateKey,
	trustedSubnet net.IPNet,
) Server {
	router := NewRouter(str, hasher, key, trustedSubnet)
	return &http.Server{Addr: addr, Handler: router}
}

// NewRouter создает новый роутер
func NewRouter(str storage.MetricsStorage, hasher hash.Hash, key *rsa.PrivateKey, trustedSubnet net.IPNet) *mux.Router {
	homeHandler := http_handlers.NewHomeHandler(str)
	metricsController := http_handlers.NewMetricsController(str, hasher)

	r := mux.NewRouter()
	r.Use(middlewares.GZIPEncode)
	r.Use(middlewares.NewDecryptMiddleware(key).Handle)
	if trustedSubnet.IP != nil {
		r.Use(middlewares.NewTrustedIPsMiddleware(trustedSubnet).Handle)
	}
	r.Handle("/", http.HandlerFunc(homeHandler.Home)).Methods(http.MethodGet, http.MethodHead)
	r.Handle("/update/", http.HandlerFunc(metricsController.UpdateMetricJSON)).Methods(http.MethodPost)
	r.Handle("/updates/", http.HandlerFunc(metricsController.UpdateMetricsBatch)).Methods(http.MethodPost)
	r.Handle("/update/{type}/{name}/{value}", http.HandlerFunc(metricsController.UpdateMetric)).Methods(http.MethodPost)
	r.Handle("/value/", http.HandlerFunc(metricsController.GetMetricJSON)).Methods(http.MethodPost)
	r.Handle("/value/{type}/{name}", http.HandlerFunc(metricsController.GetMetric)).Methods(http.MethodGet, http.MethodHead)
	r.Handle("/ping", pingHandler(str)).Methods(http.MethodGet, http.MethodHead)

	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	return r
}
