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

func TestMetrics_Sign(t *testing.T) {
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
