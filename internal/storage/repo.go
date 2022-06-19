package storage

type Repo interface {
	AddCounterMetric(string, int64)
	AddGaugeMetric(string, float64)
	GetCounter(string) (int64, error)
	GetGauge(string) (float64, error)
	AsJson() MetricStorage
}

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type MetricStorage struct {
	Metrics []Metric
}
