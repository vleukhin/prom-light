package internal

import (
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"hash"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/vleukhin/prom-light/internal/handlers"
	"github.com/vleukhin/prom-light/internal/storage"
)

type MetricsServer struct {
	cfg    *ServerConfig
	str    storage.MetricsStorage
	hasher hash.Hash
}

type gzipWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func NewMetricsServer(config *ServerConfig) (MetricsServer, error) {
	var err error
	var str storage.MetricsStorage

	switch true {
	case config.DSN != "":
		str, err = storage.NewDatabaseStorage(config.DSN, config.DBConnTimeout)
		if err != nil {
			return MetricsServer{}, err
		}
	case config.StoreFile != "":
		str, err = storage.NewFileStorage(config.StoreFile, config.StoreInterval, config.Restore)
		if err != nil {
			return MetricsServer{}, err
		}
	default:
		str = storage.NewMemoryStorage()
	}

	server := MetricsServer{
		config,
		str,
		nil,
	}

	err = server.migrate()
	if err != nil {
		return MetricsServer{}, err
	}

	if config.Key != "" {
		server.hasher = hmac.New(sha256.New, []byte(config.Key))
	}

	return server, nil
}

func (s MetricsServer) Run(err chan<- error) {
	log.Println("Metrics server listen at: " + s.cfg.Addr)
	err <- http.ListenAndServe(s.cfg.Addr, NewRouter(s.str, s.hasher))
}

func (s MetricsServer) Stop() error {
	return s.str.ShutDown(context.TODO())
}

func NewRouter(str storage.MetricsStorage, hasher hash.Hash) *mux.Router {
	homeHandler := handlers.NewHomeHandler(str)
	updateHandler := handlers.NewUpdateMetricHandler(str)
	updateJSONHandler := handlers.NewUpdateMetricJSONHandler(str, hasher)
	updateBatchHandler := handlers.NewUpdateMetricsBatchHandler(str, hasher)
	getHandler := handlers.NewGetMetricHandler(str)
	getJSONHandler := handlers.NewGetMetricJSONHandler(str, hasher)

	r := mux.NewRouter()
	r.Use(gzipEncode)
	r.Handle("/", homeHandler).Methods(http.MethodGet, http.MethodHead)
	r.Handle("/update/", updateJSONHandler).Methods(http.MethodPost)
	r.Handle("/updates/", updateBatchHandler).Methods(http.MethodPost)
	r.Handle("/update/{type}/{name}/{value}", updateHandler).Methods(http.MethodPost)
	r.Handle("/value/", getJSONHandler).Methods(http.MethodPost)
	r.Handle("/value/{type}/{name}", getHandler).Methods(http.MethodGet, http.MethodHead)
	r.Handle("/ping", pingHandler(str)).Methods(http.MethodGet, http.MethodHead)

	return r
}

func gzipEncode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
		if err != nil {
			log.Println("Failed to create gzip writer: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func(gz *gzip.Writer) {
			err := gz.Close()
			if err != nil {
				log.Println("Failed to close gzip writer: " + err.Error())
			}
		}(gz)

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
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

func (s MetricsServer) migrate() error {
	store, ok := s.str.(*storage.DatabaseStorage)
	if !ok {
		return nil
	}

	return store.Migrate(context.TODO())
}
