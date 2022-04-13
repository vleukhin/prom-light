package internal

import (
	"compress/gzip"
	"log"
	"net/http"
	"strings"

	handlers2 "github.com/vleukhin/prom-light/internal/handlers"
	storage2 "github.com/vleukhin/prom-light/internal/storage"

	"github.com/gorilla/mux"
)

type MetricsServer struct {
	cfg *ServerConfig
	str storage2.MetricsStorage
}

type gzipWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func NewMetricsServer(cfg *ServerConfig) (MetricsServer, error) {
	var err error
	var str storage2.MetricsStorage

	if cfg.StoreFile == "" {
		str = storage2.NewMemoryStorage()
	} else {
		str, err = storage2.NewFileStorage(cfg.StoreFile, cfg.StoreInterval, cfg.Restore)
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

func NewRouter(str storage2.MetricsStorage) *mux.Router {
	homeHandler := handlers2.NewHomeHandler(str)
	updateHandler := handlers2.NewUpdateMetricHandler(str)
	updateJSONHandler := handlers2.NewUpdateMetricJSONHandler(str)
	getHandler := handlers2.NewGetMetricHandler(str)
	getJSONHandler := handlers2.NewGetMetricJSONHandler(str)

	r := mux.NewRouter()
	r.Use(gzipEncode)
	r.Handle("/", homeHandler).Methods(http.MethodGet, http.MethodHead)
	r.Handle("/update/", updateJSONHandler).Methods(http.MethodPost)
	r.Handle("/update/{type}/{name}/{value}", updateHandler).Methods(http.MethodPost)
	r.Handle("/value/", getJSONHandler).Methods(http.MethodPost)
	r.Handle("/value/{type}/{name}", getHandler).Methods(http.MethodGet, http.MethodHead)

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
