package server

import (
	"github.com/stretchr/testify/require"
	"github.com/vleukhin/prom-light/cmd/server/storage"
	"github.com/vleukhin/prom-light/internal"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
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
		gaugeValue   internal.Gauge
		counterValue internal.Counter
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
			name: "POST none value to coutner",
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
	testServer := httptest.NewServer(NewRouter(mockStorage))
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
		gauges   map[string]internal.Gauge
		counters map[string]internal.Counter
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
				gauges: map[string]internal.Gauge{"test": 10.25},
			},
			want: want{
				code:  200,
				value: "10.250000",
			},
		},
		{
			name: "GET valid counter",
			request: requestOptions{
				URI:    "/value/counter/test",
				method: http.MethodGet,
			},
			metrics: storedMetrics{
				counters: map[string]internal.Counter{"test": 99},
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
	testServer := httptest.NewServer(NewRouter(mockStorage))
	defer testServer.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for name, value := range tt.metrics.gauges {
				mockStorage.StoreGauge(name, value)
			}
			for name, value := range tt.metrics.counters {
				mockStorage.StoreCounter(name, value)
			}

			req, err := http.NewRequest(tt.request.method, testServer.URL+tt.request.URI, nil)
			require.NoError(t, err)

			response, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer response.Body.Close()
			require.Equal(t, tt.want.code, response.StatusCode)

			if tt.want.code == http.StatusOK {
				respBody, err := ioutil.ReadAll(response.Body)
				require.NoError(t, err)

				require.Equal(t, tt.want.value, string(respBody))
			}
		})
	}
}
