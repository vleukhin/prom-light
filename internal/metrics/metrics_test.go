package metrics

import (
	"crypto/hmac"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeCounterPointer(i int) *Counter {
	c := Counter(i)
	return &c
}
func makeGaugePointer(i float64) *Gauge {
	g := Gauge(i)
	return &g
}

func TestMetric_Sign(t *testing.T) {
	tests := []struct {
		name     string
		metric   Metric
		wantHash string
	}{
		{
			name: "Counter",
			metric: Metric{
				Name:  "TestCounter",
				Type:  CounterTypeName,
				Delta: makeCounterPointer(100),
				Value: nil,
			},
			wantHash: "e53ee57da9924b1f68c305121e89c39a4104109ac37bed219a66c13920dc3d72",
		},
		{
			name: "Gauge",
			metric: Metric{
				Name:  "TestGauge",
				Type:  GaugeTypeName,
				Delta: nil,
				Value: makeGaugePointer(125.3444442),
			},
			wantHash: "483c1e0d3e3b33b9426863bda45a142581c837db875f2d4087ab7b74d76a3c9f",
		},
	}

	hasher := hmac.New(sha256.New, []byte("test-key"))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.metric.Sign(hasher)
			assert.Equal(t, tt.wantHash, tt.metric.Hash)
		})
	}
}

func TestMetric_String(t *testing.T) {
	tests := []struct {
		name   string
		metric Metric
		str    string
	}{
		{
			name:   "round gauge",
			metric: MakeGaugeMetric("test", 10),
			str:    "10.000",
		},
		{
			name:   "gauge",
			metric: MakeGaugeMetric("test", 10.135496),
			str:    "10.135",
		},
		{
			name:   "counter",
			metric: MakeCounterMetric("test", 10),
			str:    "10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.str, tt.metric.String())
		})
	}
}

func BenchmarkSign(b *testing.B) {
	var metricsData = Metrics{
		MakeCounterMetric("Counter1", 0),
		MakeCounterMetric("Counter2", 12312),
		MakeCounterMetric("Counter3", 4444),
		MakeGaugeMetric("Gauge1", 5.5),
		MakeGaugeMetric("Gauge2", 0),
		MakeGaugeMetric("Gauge3", -8),
	}
	hasher := hmac.New(sha256.New, []byte("test-key"))

	for i := 0; i < b.N; i++ {
		metricsData.Sign(hasher)
	}
}
