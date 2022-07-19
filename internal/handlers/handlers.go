package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v4"
	"github.com/zamibaru89/devops-metrics/internal/config"
	"github.com/zamibaru89/devops-metrics/internal/functions"
	"github.com/zamibaru89/devops-metrics/internal/storage"
	"log"
	"net/http"
	"strconv"
)

func ReceiveMetric(config config.ServerConfig, storage storage.Repo) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")
		switch metricType {
		case "gauge":
			Value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.AddGaugeMetric(metricName, Value)
			if config.StoreInterval == 0 && config.DSN == "" {
				err := functions.SaveMetricToDisk(config, storage)
				if err != nil {
					return
				}
			}
		case "counter":
			Value, err := strconv.ParseInt(metricValue, 0, 64)
			if err != nil {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.AddCounterMetric(metricName, Value)
			if config.StoreInterval == 0 && config.DSN == "" {
				err := functions.SaveMetricToDisk(config, storage)
				if err != nil {
					return
				}
			}

		default:

			w.WriteHeader(501)

		}

	}
}

func ValueOfMetric(storage storage.Repo) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		metricName := chi.URLParam(r, "metricName")
		metricType := chi.URLParam(r, "metricType")

		switch metricType {
		case "gauge":
			value, err := storage.GetGauge(metricName)
			if err != nil {
				fmt.Println(value, err)
				w.WriteHeader(http.StatusNotFound)

			}
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintln(w, value)
		case "counter":
			value, err := storage.GetCounter(metricName)
			if err != nil {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusNotFound)

			}
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintln(w, value)
		default:
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
		}

	}
}
func ListMetrics(storage storage.Repo) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		json, err := json.Marshal(storage.AsMetric())
		if err != nil {
			log.Println(err)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, string(json))
	}
}

func ReceiveMetricJSON(config config.ServerConfig, st storage.Repo) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var m storage.Metric
		err := json.NewDecoder(r.Body).Decode(&m)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, m)

			return
		}

		switch m.MType {
		case "gauge":
			if m.Value == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
			} else {
				st.AddGaugeMetric(m.ID, *m.Value)
				w.Header().Set("Content-Type", "application/json")
				render.JSON(w, r, m)
				if config.StoreInterval == 0 && config.DSN == "" {
					functions.SaveMetricToDisk(config, st)

				}
			}
		case "counter":
			if m.Delta == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)

				render.JSON(w, r, m)

			} else {
				st.AddCounterMetric(m.ID, *m.Delta)
				w.Header().Set("Content-Type", "application/json")
				render.JSON(w, r, m)

				if config.StoreInterval == 0 && config.DSN == "" {
					functions.SaveMetricToDisk(config, st)

				}
			}
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
		}

	}
}

func ValueOfMetricJSON(config config.ServerConfig, st storage.Repo) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var m storage.Metric
		err := json.NewDecoder(r.Body).Decode(&m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch m.MType {
		case "counter":
			value, err := st.GetCounter(m.ID)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			m.Delta = &value
			message := fmt.Sprintf("%s:counter:%d", m.ID, value)
			if config.Key != "" {
				m.Hash = functions.CreateHash(message, []byte(config.Key))
			} else {
				m.Hash = ""
			}
			w.Header().Set("Content-Type", "application/json")
			render.JSON(w, r, m)
		case "gauge":
			value, err := st.GetGauge(m.ID)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			m.Value = &value
			message := fmt.Sprintf("%s:gauge:%f", m.ID, value)
			if config.Key != "" {
				m.Hash = functions.CreateHash(message, []byte(config.Key))
			} else {
				m.Hash = ""
			}
			w.Header().Set("Content-Type", "application/json")
			render.JSON(w, r, m)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func PingDB(config config.ServerConfig, conn *pgx.Conn) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if config.DSN == "" {
			w.WriteHeader(http.StatusNotAcceptable)
		}

		err := conn.Ping(context.Background())
		if err != nil {
			log.Printf("Unable to connect to DB: %v\n\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

	}
}

func ReceiveMetricsJSON(config config.ServerConfig, st storage.Repo) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var m []storage.Metric
		err := json.NewDecoder(r.Body).Decode(&m)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, m)
			fmt.Println(err)

			return
		}
		st.AddMetrics(m)
		if config.StoreInterval == 0 && config.DSN == "" {
			functions.SaveMetricToDisk(config, st)

		}

	}
}
