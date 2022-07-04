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
    metric_value DOUBLE PRECISION,
    metric_hash VARCHAR(100)
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

	query := "SELECT metric_value FROM metrics WHERE metric_id=$1"

	result, err := conn.Query(context.Background(), query, metricName)
	if err != nil {
		return 0, err
	}
	defer result.Close()
	result.Scan(&gauge)

	return gauge, nil
}

func (p *PostgresStorage) GetCounter(metricName string) (int64, error) {
	var counter int64
	conn, err := pgx.Connect(context.Background(), p.DSN)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close(context.Background())

	query := "SELECT metric_delta FROM metrics WHERE metric_id=$1"

	result, err := conn.Query(context.Background(), query, metricName)
	if err != nil {
		return 0, err
	}
	defer result.Close()
	result.Scan(&counter)

	return counter, nil
}

func (p *PostgresStorage) AsMetric() MetricStorage {
	fmt.Println("PH")
	return MetricStorage{}
}
