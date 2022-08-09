package metrics

import (
	"encoding/hex"
	"fmt"
	"hash"
)

type Gauge float64
type Counter int64

type Metric struct {
	Name  string   `json:"id"`              // имя метрики
	Type  string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *Counter `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *Gauge   `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type Metrics []Metric

const GaugeTypeName = "gauge"
const CounterTypeName = "counter"

func (m Metric) IsCounter() bool {
	return m.Type == CounterTypeName
}

func (m Metric) String() string {
	var str string
	switch m.Type {
	case GaugeTypeName:
		str = fmt.Sprintf("%.3f", *m.Value)
	case CounterTypeName:
		str = fmt.Sprintf("%d", *m.Delta)
	default:
		str = "unknown"
	}

	return str
}

func makeHash(m Metric, hasher hash.Hash) string {
	if hasher == nil {
		return ""
	}

	switch m.Type {
	case CounterTypeName:
		hasher.Write([]byte(fmt.Sprintf("%s:counter:%d", m.Name, *m.Delta)))
	case GaugeTypeName:
		hasher.Write([]byte(fmt.Sprintf("%s:gauge:%f", m.Name, *m.Value)))
	}

	defer hasher.Reset()
	return hex.EncodeToString(hasher.Sum(nil))
}

func (m *Metric) Sign(hasher hash.Hash) {
	if hasher == nil {
		return
	}
	m.Hash = makeHash(*m, hasher)
}

func (m Metric) IsValid(hasher hash.Hash) bool {
	return m.Hash == makeHash(m, hasher)
}

func (m Metrics) IsValid(hasher hash.Hash) bool {
	for _, i := range m {
		if !i.IsValid(hasher) {
			return false
		}
	}
	return true
}

func (m Metrics) Sign(hasher hash.Hash) Metrics {
	result := make(Metrics, len(m))
	for i, metric := range m {
		metric.Sign(hasher)
		result[i] = metric
	}

	return result
}

func MakeGaugeMetric(name string, value Gauge) Metric {
	return Metric{
		Name:  name,
		Type:  GaugeTypeName,
		Value: &value,
	}
}
func MakeCounterMetric(name string, delta Counter) Metric {
	return Metric{
		Name:  name,
		Type:  CounterTypeName,
		Delta: &delta,
	}
}
