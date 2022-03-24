package handlers

import (
	"fmt"
	"github.com/vleukhin/prom-light/internal"
	"net/http"
	"regexp"
	"strconv"
)

type UpdateMetricHandler struct {
	storage MetricsStorage
}

type MetricsStorage interface {
	StoreGauge(metricName string, value internal.Gauge)
	StoreCounter(metricName string, value internal.Counter)
}

func NewUpdateMetricHandler(storage MetricsStorage) UpdateMetricHandler {
	return UpdateMetricHandler{
		storage: storage,
	}
}

func (h UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	myExp := regexp.MustCompile("^/update/(?P<mType>\\w*)/(?P<mName>\\w*)/(?P<mValue>[.\\d]+)$")
	match := myExp.FindStringSubmatch(r.RequestURI)

	if len(match) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	params := make(map[string]string)
	for i, name := range myExp.SubexpNames() {
		if i != 0 && name != "" {
			params[name] = match[i]
		}
	}

	switch internal.MetricTypeName(params["mType"]) {
	case internal.GaugeTypeName:
		value, err := strconv.ParseFloat(params["mValue"], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Printf("Received gauge %s with value %f \n", params["mName"], value)
		h.storage.StoreGauge(params["mName"], internal.Gauge(value))
	case internal.CounterTypeName:
		value, err := strconv.ParseInt(params["mValue"], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Printf("Received counter %s with value %d \n", params["mName"], value)
		h.storage.StoreCounter(params["mName"], internal.Counter(value))
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err := w.Write([]byte("Updated"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
