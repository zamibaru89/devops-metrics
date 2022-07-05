package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/zamibaru89/devops-metrics/internal/config"
	"log"
)

type PostgresStorage struct {
	DSN string
}

func NewPostgresStorage(c config.ServerConfig) Repo {
	conn, err := pgx.Connect(context.Background(), c.DSN)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close(context.Background())
	query := `CREATE TABLE IF NOT EXISTS  metrics(
    id serial PRIMARY KEY,
    metric_id VARCHAR(50) NOT NULL UNIQUE,
    metric_type VARCHAR(50),
    metric_delta BIGINT,
    metric_value DOUBLE PRECISION
);`
	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		log.Println(err)
	}

	return &PostgresStorage{DSN: c.DSN}
}
func (p *PostgresStorage) AddCounterMetric(name string, value int64) {
	conn, err := pgx.Connect(context.Background(), p.DSN)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close(context.Background())
	query := `
		INSERT INTO metrics(
		metric_id,
		metric_type,
		metric_delta 
		)
		VALUES($1, $2, $3)
		ON CONFLICT (metric_id) DO UPDATE
		SET metric_delta=metrics.metric_delta+$3;
	`
	_, err = conn.Exec(context.Background(), query, name, "counter", value)
	if err != nil {
		log.Println(err)
	}
}

func (p *PostgresStorage) AddGaugeMetric(name string, value float64) {
	conn, err := pgx.Connect(context.Background(), p.DSN)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close(context.Background())
	query := `
		INSERT INTO metrics(
		metric_id,
		metric_type,
		metric_value
		)
		VALUES($1, $2, $3)
		ON CONFLICT (metric_id) DO UPDATE
		SET metric_value=$3;
	`
	_, err = conn.Exec(context.Background(), query, name, "gauge", value)
	if err != nil {
		log.Println(err)
	}
}

func (p *PostgresStorage) GetGauge(metricName string) (float64, error) {

	var gauge float64
	conn, err := pgx.Connect(context.Background(), p.DSN)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close(context.Background())

	query := "SELECT metric_value FROM metrics WHERE metric_id=$1;"

	result, err := conn.Query(context.Background(), query, metricName)

	if err != nil {
		return 0, err
	}
	defer result.Close()
	for result.Next() {
		fmt.Println("scan")
		err = result.Scan(&gauge)
		if err != nil {
			return 0, err
		}
	}

	return gauge, nil
}

func (p *PostgresStorage) GetCounter(metricName string) (int64, error) {
	var counter int64
	conn, err := pgx.Connect(context.Background(), p.DSN)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close(context.Background())

	query := "SELECT metric_delta FROM metrics WHERE metric_id=$1;"

	result, err := conn.Query(context.Background(), query, metricName)

	if err != nil {
		return 0, err
	}
	defer result.Close()
	for result.Next() {
		err = result.Scan(&counter)
		if err != nil {
			return 0, err
		}
	}

	return counter, nil
}

func (p *PostgresStorage) AsMetric() MetricStorage {
	var metrics MetricStorage
	conn, err := pgx.Connect(context.Background(), p.DSN)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close(context.Background())
	query := "SELECT metric_id, metric_type, metric_delta, metric_value FROM metrics"

	result, err := conn.Query(context.Background(), query)
	if err != nil {
		log.Println(err)
	}
	defer result.Close()

	for result.Next() {
		var metricID, metricType string
		var metricDelta *int64
		var metricValue *float64
		err := result.Scan(&metricID, &metricType, &metricDelta, &metricValue)
		if err != nil {
			log.Println(err)
		}

		metrics.Metrics = append(metrics.Metrics, Metric{
			ID:    metricID,
			MType: metricType,
			Delta: metricDelta,
			Value: metricValue,
		})
	}

	return metrics
}

func (p *PostgresStorage) AddMetrics(metrics MetricStorage) {
	conn, err := pgx.Connect(context.Background(), p.DSN)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close(context.Background())
	tran, _ := conn.Begin(context.Background())
	defer tran.Rollback(context.Background())
	for i := range metrics.Metrics {

		query := `
		INSERT INTO metrics(
		metric_id,
		metric_type,
		metric_value, 
		metric_delta
		)
		VALUES($1, $2, $3, $4)
		ON CONFLICT (metric_id) DO UPDATE
		SET metric_value=$3, metric_delta=metrics.metric_delta+$4;
	`
		_, err = conn.Exec(context.Background(), query, metrics.Metrics[i].ID, metrics.Metrics[i].MType, metrics.Metrics[i].Value, metrics.Metrics[i].Delta)
		if err != nil {
			log.Println(err)
		}

	}
	tran.Commit(context.Background())

}
