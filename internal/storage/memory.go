package storage

import (
	"errors"
	"sync"
)

type MemoryStorage struct {
	GaugeMetric   map[string]float64
	GaugeMutex    sync.Mutex
	CounterMetric map[string]int64
	CounterMutex  sync.Mutex
}

func NewMemoryStorage() Repo {
	return &MemoryStorage{
		GaugeMetric:   make(map[string]float64),
		CounterMetric: make(map[string]int64),
	}
}
func (m *MemoryStorage) AddCounterMetric(name string, value int64) {
	if len(m.CounterMetric) == 0 {
		m.CounterMetric = make(map[string]int64)
	}
	m.CounterMetric[name] += value
}

func (m *MemoryStorage) AddGaugeMetric(name string, value float64) {
	if len(m.GaugeMetric) == 0 {
		m.GaugeMetric = make(map[string]float64)
	}
	m.GaugeMetric[name] = value
}

func (m *MemoryStorage) GetGauge(metricName string) (float64, error) {

	if value, ok := m.GaugeMetric[metricName]; ok {
		return value, nil
	}
	return 0, errors.New("not found")

}

func (m *MemoryStorage) GetCounter(metricName string) (int64, error) {

	if value, ok := m.CounterMetric[metricName]; ok {
		return value, nil
	}
	return 0, errors.New("not Found")
}

func (m *MemoryStorage) AsJSON() MetricStorage {
	var metrics MetricStorage
	for id, value := range m.GaugeMetric {
		value := value
		metrics.Metrics = append(metrics.Metrics, Metric{
			ID:    id,
			MType: "gauge",
			Value: &value,
		})
	}
	for id, delta := range m.CounterMetric {
		delta := delta
		metrics.Metrics = append(metrics.Metrics, Metric{
			ID:    id,
			MType: "counter",
			Delta: &delta,
		})
	}

	return metrics
}
