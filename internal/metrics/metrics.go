package metrics

import (
	"encoding/hex"
	"fmt"
	"hash"
)

type Gauge float64
type Counter int64

// Metric определяем метрику
type Metric struct {
	// Name имя метрики
	Name string `json:"id"`
	// Type параметр, принимающий значение gauge или counter
	Type string `json:"type"`
	// Delta значение метрики в случае передачи counter
	Delta *Counter `json:"delta,omitempty"`
	// Value значение метрики в случае передачи gauge
	Value *Gauge `json:"value,omitempty"`
	// Hash значение хеш-функции
	Hash string `json:"hash,omitempty"`
}

// Metrics слайс метрик
type Metrics []Metric

// Строковое представление типов метрик
const (
	GaugeTypeName   = "gauge"
	CounterTypeName = "counter"
)

// IsCounter проверяет является ли метрика счетчиком
func (m Metric) IsCounter() bool {
	return m.Type == CounterTypeName
}

// String выдает строковое представление метрики
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

// makeHash вычисляем хэш для метрики
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

// Sign подписывает метрику с помощью переданного хэшера
func (m *Metric) Sign(hasher hash.Hash) {
	if hasher == nil {
		return
	}
	m.Hash = makeHash(*m, hasher)
}

// IsValid проверяет валидность подписи метрики
func (m Metric) IsValid(hasher hash.Hash) bool {
	return m.Hash == makeHash(m, hasher)
}

// IsValid проверяет валидность всех метрик в слайсе
func (m Metrics) IsValid(hasher hash.Hash) bool {
	for _, i := range m {
		if !i.IsValid(hasher) {
			return false
		}
	}
	return true
}

// Sign подписывает все метрики в слайсе
func (m Metrics) Sign(hasher hash.Hash) Metrics {
	result := make(Metrics, len(m))
	for i, metric := range m {
		metric.Sign(hasher)
		result[i] = metric
	}

	return result
}

// MakeGaugeMetric создает метрику типа gauge
func MakeGaugeMetric(name string, value Gauge) Metric {
	return Metric{
		Name:  name,
		Type:  GaugeTypeName,
		Value: &value,
	}
}

// MakeCounterMetric создает метрику типа counter
func MakeCounterMetric(name string, delta Counter) Metric {
	return Metric{
		Name:  name,
		Type:  CounterTypeName,
		Delta: &delta,
	}
}
