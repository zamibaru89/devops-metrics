package storage

type Repo interface {
	AddCounterMetric(string, int64) error
	AddGaugeMetric(string, float64) error
	GetCounter(string) (int64, error)
	GetGauge(string) (float64, error)
	AsMetric() MetricStorage
	AddMetrics(metrics []Metric) error
}

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type MetricStorage struct {
	Metrics []Metric
}
