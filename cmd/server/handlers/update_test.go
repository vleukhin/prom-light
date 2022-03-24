package handlers

import (
	"github.com/stretchr/testify/require"
	"github.com/vleukhin/prom-light/cmd/server/storage"
	"github.com/vleukhin/prom-light/internal"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateMetricHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code         int
		success      bool
		metricName   string
		gaugeValue   float64
		counterValue int64
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name    string
		storage MetricsStorage
		request *http.Request
		want    want
	}{
		{
			name:    "GET /undefined",
			request: httptest.NewRequest(http.MethodGet, "/undefined", nil),
			want: want{
				code:    405,
				success: false,
			},
		},
		{
			name:    "GET valid metric",
			request: httptest.NewRequest(http.MethodGet, "/update/gauge/test/1", nil),
			want: want{
				code:    405,
				success: false,
			},
		},
		{
			name:    "POST valid gauge metric",
			request: httptest.NewRequest(http.MethodPost, "/update/gauge/testGauge/1.25", nil),
			want: want{
				code:       200,
				success:    true,
				metricName: "testGauge",
				gaugeValue: 1.25,
			},
		},
		{
			name:    "POST valid counter metric",
			request: httptest.NewRequest(http.MethodPost, "/update/counter/testCounter/5", nil),
			want: want{
				code:         200,
				success:      true,
				metricName:   "testCounter",
				counterValue: 5,
			},
		},
		{
			name:    "POST none value to gauge",
			request: httptest.NewRequest(http.MethodPost, "/update/gauge/testCounter/none", nil),
			want: want{
				code:    400,
				success: false,
			},
		},
		{
			name:    "POST none value to coutner",
			request: httptest.NewRequest(http.MethodPost, "/update/counter/testCounter/none", nil),
			want: want{
				code:    400,
				success: false,
			},
		},
		{
			name:    "POST unknown metric type",
			request: httptest.NewRequest(http.MethodPost, "/update/unknown/testCounter/none", nil),
			want: want{
				code:    501,
				success: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := storage.NewMockStorage(t)
			w := httptest.NewRecorder()
			h := UpdateMetricHandler{
				storage: mockStorage,
			}
			h.ServeHTTP(w, tt.request)
			response := w.Result()
			defer response.Body.Close()
			require.Equal(t, tt.want.code, response.StatusCode)

			if tt.want.success {
				if tt.want.gaugeValue != 0 {
					mockStorage.AssertGaugeStoredWithValue(tt.want.metricName, internal.Gauge(tt.want.gaugeValue))
				}
				if tt.want.counterValue != 0 {
					mockStorage.AssertCounterStoredWithValue(tt.want.metricName, internal.Counter(tt.want.counterValue))
				}
			}
		})
	}
}
