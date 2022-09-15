package server

import (
	"crypto/rsa"
	"github.com/gorilla/mux"
	"github.com/vleukhin/prom-light/internal/handlers"
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
) *http.Server {
	router := NewRouter(str, hasher, key, trustedSubnet)
	return &http.Server{Addr: addr, Handler: router}
}

// NewRouter создает новый роутер
func NewRouter(str storage.MetricsStorage, hasher hash.Hash, key *rsa.PrivateKey, trustedSubnet net.IPNet) *mux.Router {
	homeHandler := handlers.NewHomeHandler(str)
	updateHandler := handlers.NewUpdateMetricHandler(str)
	updateJSONHandler := handlers.NewUpdateMetricJSONHandler(str, hasher)
	updateBatchHandler := handlers.NewUpdateMetricsBatchHandler(str, hasher)
	getHandler := handlers.NewGetMetricHandler(str)
	getJSONHandler := handlers.NewGetMetricJSONHandler(str, hasher)

	r := mux.NewRouter()
	r.Use(middlewares.GZIPEncode)
	r.Use(middlewares.NewDecryptMiddleware(key).Handle)
	if trustedSubnet.IP != nil {
		r.Use(middlewares.NewTrustedIPsMiddleware(trustedSubnet).Handle)
	}
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
