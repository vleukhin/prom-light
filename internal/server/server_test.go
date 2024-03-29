package server

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vleukhin/prom-light/internal/config"

	"github.com/stretchr/testify/require"

	"github.com/vleukhin/prom-light/internal/metrics"
	"github.com/vleukhin/prom-light/internal/storage"
)

type requestOptions struct {
	URI    string
	method string
}

func TestUpdateMetricHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code         int
		success      bool
		metricName   string
		gaugeValue   metrics.Gauge
		counterValue metrics.Counter
	}

	tests := []struct {
		name    string
		request requestOptions
		want    want
	}{
		{
			name: "GET /undefined",
			request: requestOptions{
				URI:    "/undefined",
				method: http.MethodGet,
			},
			want: want{
				code:    404,
				success: false,
			},
		},
		{
			name: "GET valid metric",
			request: requestOptions{
				URI:    "/update/gauge/test/1",
				method: http.MethodGet,
			},
			want: want{
				code:    405,
				success: false,
			},
		},
		{
			name: "POST valid gauge metric",
			request: requestOptions{
				URI:    "/update/gauge/testGauge/1.25",
				method: http.MethodPost,
			},
			want: want{
				code:       200,
				success:    true,
				metricName: "testGauge",
				gaugeValue: 1.25,
			},
		},
		{
			name: "POST valid counter metric",
			request: requestOptions{
				URI:    "/update/counter/testCounter/5",
				method: http.MethodPost,
			},
			want: want{
				code:         200,
				success:      true,
				metricName:   "testCounter",
				counterValue: 5,
			},
		},
		{
			name: "POST none value to gauge",
			request: requestOptions{
				URI:    "/update/gauge/testCounter/none",
				method: http.MethodPost,
			},
			want: want{
				code:    400,
				success: false,
			},
		},
		{
			name: "POST none value to counter",
			request: requestOptions{
				URI:    "/update/counter/testCounter/none",
				method: http.MethodPost,
			},
			want: want{
				code:    400,
				success: false,
			},
		},
		{
			name: "POST unknown metric type",
			request: requestOptions{
				URI:    "/update/unknown/testCounter/none",
				method: http.MethodPost,
			},
			want: want{
				code:    501,
				success: false,
			},
		},
	}

	mockStorage := storage.NewMockStorage()
	testServer := httptest.NewServer(NewRouter(mockStorage, nil, nil, net.IPNet{}))
	defer testServer.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.request.method, testServer.URL+tt.request.URI, nil)
			require.NoError(t, err)

			response, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer response.Body.Close()
			require.Equal(t, tt.want.code, response.StatusCode)

			if tt.want.success {
				if tt.want.gaugeValue != 0 {
					mockStorage.AssertGaugeStoredWithValue(t, tt.want.metricName, tt.want.gaugeValue)
				}
				if tt.want.counterValue != 0 {
					mockStorage.AssertCounterStoredWithValue(t, tt.want.metricName, tt.want.counterValue)
				}
			}
		})
	}
}

func TestGetMetricHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code  int
		value string
	}
	type storedMetrics struct {
		gauges   map[string]metrics.Gauge
		counters map[string]metrics.Counter
	}
	tests := []struct {
		name    string
		metrics storedMetrics
		request requestOptions
		want    want
	}{
		{
			name: "GET /undefined",
			request: requestOptions{
				URI:    "/undefined",
				method: http.MethodGet,
			},
			want: want{
				code: 404,
			},
		},
		{
			name: "GET valid gauge",
			request: requestOptions{
				URI:    "/value/gauge/test",
				method: http.MethodGet,
			},
			metrics: storedMetrics{
				gauges: map[string]metrics.Gauge{"test": 10.25},
			},
			want: want{
				code:  200,
				value: "10.250",
			},
		},
		{
			name: "GET valid counter",
			request: requestOptions{
				URI:    "/value/counter/test",
				method: http.MethodGet,
			},
			metrics: storedMetrics{
				counters: map[string]metrics.Counter{"test": 99},
			},
			want: want{
				code:  200,
				value: "99",
			},
		},
		{
			name: "GET unknown metric",
			request: requestOptions{
				URI:    "/value/vector/test",
				method: http.MethodGet,
			},
			want: want{
				code: 501,
			},
		},
	}

	mockStorage := storage.NewMockStorage()
	testServer := httptest.NewServer(NewRouter(mockStorage, nil, nil, net.IPNet{}))
	defer testServer.Close()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for name, value := range tt.metrics.gauges {
				_ = mockStorage.SetGauge(ctx, name, value)
			}
			for name, value := range tt.metrics.counters {
				_ = mockStorage.IncCounter(ctx, name, value)
			}

			req, err := http.NewRequest(tt.request.method, testServer.URL+tt.request.URI, nil)
			require.NoError(t, err)

			response, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer response.Body.Close()
			require.Equal(t, tt.want.code, response.StatusCode)

			if tt.want.code == http.StatusOK {
				respBody, err := io.ReadAll(response.Body)
				require.NoError(t, err)

				require.Equal(t, tt.want.value, string(respBody))
			}
		})
	}
}

func TestHomeHandler_ServeHTTP(t *testing.T) {
	mockStorage := storage.NewMockStorage()
	_ = mockStorage.IncCounter(context.Background(), "foo", 1)
	testServer := httptest.NewServer(NewRouter(mockStorage, nil, nil, net.IPNet{}))
	req, err := http.NewRequest(http.MethodGet, testServer.URL, nil)
	require.NoError(t, err)

	response, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)
}

func TestUpdateMetricJSONHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code   int
		metric metrics.Metric
	}

	var testCounter metrics.Counter = 5
	var testGauge metrics.Gauge = 5.5

	var tests = []struct {
		name    string
		payload []byte
		want    want
	}{
		{
			name:    "Bad request",
			payload: []byte("test"),
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:    "Single counter",
			payload: []byte(`{"id":"TestCounter","type":"counter","delta":5}`),
			want: want{
				code: http.StatusOK,
				metric: metrics.Metric{
					Name:  "TestCounter",
					Type:  metrics.CounterTypeName,
					Delta: &testCounter,
				},
			},
		},
		{
			name:    "Single gauge",
			payload: []byte(`{"id":"TestGauge","type":"gauge","value":5.5}`),
			want: want{
				code: http.StatusOK,
				metric: metrics.Metric{
					Name:  "TestGauge",
					Type:  metrics.GaugeTypeName,
					Value: &testGauge,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := storage.NewMockStorage()
			testServer := httptest.NewServer(NewRouter(mockStorage, nil, nil, net.IPNet{}))
			defer testServer.Close()

			req, err := http.NewRequest(http.MethodPost, testServer.URL+"/update/", bytes.NewBuffer(tt.payload))
			require.NoError(t, err)

			response, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer response.Body.Close()
			require.Equal(t, tt.want.code, response.StatusCode)

			if tt.want.code == http.StatusOK {
				switch tt.want.metric.Type {
				case metrics.GaugeTypeName:
					mockStorage.AssertGaugeStoredWithValue(t, tt.want.metric.Name, *tt.want.metric.Value)
				case metrics.CounterTypeName:
					mockStorage.AssertCounterStoredWithValue(t, tt.want.metric.Name, *tt.want.metric.Delta)
				}
			}
		})
	}
}

func TestBatchUpdateMetricJSONHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code    int
		metrics metrics.Metrics
	}

	var testCounter metrics.Counter = 5
	var testGauge metrics.Gauge = 5.5

	var tests = []struct {
		name    string
		payload []byte
		want    want
	}{
		{
			name:    "Bad request",
			payload: []byte("test"),
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:    "Batch",
			payload: []byte(`[{"id":"TestCounter","type":"counter","delta":5},{"id":"TestGauge","type":"gauge","value":5.5}]`),
			want: want{
				code: http.StatusOK,
				metrics: metrics.Metrics{
					{
						Name:  "TestCounter",
						Type:  metrics.CounterTypeName,
						Delta: &testCounter,
					},
					{
						Name:  "TestGauge",
						Type:  metrics.GaugeTypeName,
						Value: &testGauge,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := storage.NewMockStorage()
			testServer := httptest.NewServer(NewRouter(mockStorage, nil, nil, net.IPNet{}))
			defer testServer.Close()

			req, err := http.NewRequest(http.MethodPost, testServer.URL+"/updates/", bytes.NewBuffer(tt.payload))
			require.NoError(t, err)

			response, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer response.Body.Close()
			require.Equal(t, tt.want.code, response.StatusCode)

			if tt.want.code == http.StatusOK {
				for _, m := range tt.want.metrics {
					switch m.Type {
					case metrics.GaugeTypeName:
						mockStorage.AssertGaugeStoredWithValue(t, m.Name, *m.Value)
					case metrics.CounterTypeName:
						mockStorage.AssertCounterStoredWithValue(t, m.Name, *m.Delta)
					}
				}
			}
		})
	}
}

func TestGetMetricJSONHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code     int
		response string
	}
	type storedMetrics struct {
		gauges   map[string]metrics.Gauge
		counters map[string]metrics.Counter
	}
	tests := []struct {
		name    string
		metrics storedMetrics
		payload []byte
		want    want
	}{
		{
			name:    "Bad request",
			payload: []byte("test"),
			want: want{
				code: 404,
			},
		},
		{
			name:    "Get counter",
			payload: []byte(`{"id":"TestCounter","type":"counter"}`),
			metrics: storedMetrics{
				counters: map[string]metrics.Counter{"TestCounter": 99},
			},
			want: want{
				code:     200,
				response: `{"id":"TestCounter","type":"counter","delta":99}`,
			},
		},
		{
			name:    "Get gauge",
			payload: []byte(`{"id":"TestGauge","type":"gauge"}`),
			metrics: storedMetrics{
				gauges: map[string]metrics.Gauge{"TestGauge": 99.99},
			},
			want: want{
				code:     200,
				response: `{"id":"TestGauge","type":"gauge","value":99.99}`,
			},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := storage.NewMockStorage()
			testServer := httptest.NewServer(NewRouter(mockStorage, nil, nil, net.IPNet{}))
			defer testServer.Close()

			for name, value := range tt.metrics.gauges {
				_ = mockStorage.SetGauge(ctx, name, value)
			}
			for name, value := range tt.metrics.counters {
				_ = mockStorage.IncCounter(ctx, name, value)
			}

			req, err := http.NewRequest(http.MethodPost, testServer.URL+"/value/", bytes.NewBuffer(tt.payload))
			require.NoError(t, err)

			response, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer response.Body.Close()
			require.Equal(t, tt.want.code, response.StatusCode)

			if tt.want.code == http.StatusOK {
				respBody, err := io.ReadAll(response.Body)
				require.NoError(t, err)

				require.Equal(t, tt.want.response, string(respBody))
			}
		})
	}
}

func TestServer_Start(t *testing.T) {
	server, err := NewApp(&config.ServerConfig{
		Addr:      "localhost:9999",
		StoreFile: "/tmp/server_test",
		Protocol:  config.ProtocolHTTP,
	})
	if err != nil {
		t.Error(err)
		return
	}
	ch := make(chan error)
	go server.Run(ch)
	time.Sleep(100 * time.Millisecond)
	select {
	case err := <-ch:
		t.Error(err)
	default:
		err = server.Stop(context.Background())
		if err != nil {
			t.Error(err)
		}
	}
}
