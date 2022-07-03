//TODO:
//Обсудить на 1 на 1 как реализовать
//1) Фикс дупликацию с хендлером и вынос бизнес логики
//2) Реализация апи клиента
//3) config.GetAgentConfig() и варианты реализации

package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/zamibaru89/devops-metrics/internal/config"
	"github.com/zamibaru89/devops-metrics/internal/middleware"
	"github.com/zamibaru89/devops-metrics/internal/storage"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var Server = storage.NewMemoryStorage()
var ServerConfig = config.ServerConfig{}

func receiveMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	if metricType == "gauge" {
		Value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		Server.AddGaugeMetric(metricName, Value)
		if ServerConfig.StoreInterval == 0 {
			err := SaveMetricToDisk(ServerConfig, Server)
			if err != nil {
				return
			}
		}

	} else if metricType == "counter" {

		Value, err := strconv.ParseInt(metricValue, 0, 64)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		Server.AddCounterMetric(metricName, Value)
		if ServerConfig.StoreInterval == 0 {
			err := SaveMetricToDisk(ServerConfig, Server)
			if err != nil {
				return
			}
		}

	} else {
		w.WriteHeader(501)
	}

}

func valueOfMetric(w http.ResponseWriter, r *http.Request) {

	metricName := chi.URLParam(r, "metricName")
	metricType := chi.URLParam(r, "metricType")

	switch metricType {
	case "gauge":
		value, err := Server.GetGauge(metricName)
		if err != nil {
			fmt.Println(value, err)
			w.WriteHeader(http.StatusNotFound)

		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, value)
	case "counter":
		value, err := Server.GetCounter(metricName)
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
func listMetrics(w http.ResponseWriter, r *http.Request) {
	json, err := json.Marshal(Server.AsMetric())
	if err != nil {
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintln(w, string(json))
}

func receiveMetricJSON(w http.ResponseWriter, r *http.Request) {
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
			Server.AddGaugeMetric(m.ID, *m.Value)
			w.Header().Set("Content-Type", "application/json")
			render.JSON(w, r, m)
			if ServerConfig.StoreInterval == 0 {
				SaveMetricToDisk(ServerConfig, Server)
				log.Println(m)
			}
		}
	case "counter":
		if m.Delta == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, m)
		} else {
			Server.AddCounterMetric(m.ID, *m.Delta)
			w.Header().Set("Content-Type", "application/json")
			render.JSON(w, r, m)
			if ServerConfig.StoreInterval == 0 {
				SaveMetricToDisk(ServerConfig, Server)
				log.Println(m)
			}
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
	}

}

func valueOfMetricJSON(w http.ResponseWriter, r *http.Request) {
	var m storage.Metric
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch m.MType {
	case "counter":
		value, err := Server.GetCounter(m.ID)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		m.Delta = &value
		w.Header().Set("Content-Type", "application/json")
		render.JSON(w, r, m)
	case "gauge":
		value, err := Server.GetGauge(m.ID)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		m.Value = &value
		w.Header().Set("Content-Type", "application/json")
		render.JSON(w, r, m)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
	}

}

func SaveMetricToDisk(config config.ServerConfig, m storage.Repo) error {

	filePath := config.FilePath
	fileBits := os.O_WRONLY | os.O_CREATE | os.O_TRUNC

	file, err := os.OpenFile(filePath, fileBits, 0600)
	if err != nil {
		log.Println(err)
		file.Close()
		return err
	}

	data, err := json.Marshal(m.AsMetric())
	if err != nil {
		log.Println(err)
		file.Close()
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		log.Println(err)
		file.Close()
	}

	file.Close()
	return nil
}

func RestoreMetricsFromDisk(config config.ServerConfig, r storage.Repo) storage.Repo {
	repo := r
	path := config.FilePath

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)

	}

	metrics := storage.MetricStorage{}
	json.Unmarshal(data, &metrics)

	for i := range metrics.Metrics {
		metricName := metrics.Metrics[i].ID
		if metrics.Metrics[i].MType == "counter" {
			Delta := metrics.Metrics[i].Delta
			repo.AddCounterMetric(metricName, *Delta)
		}
		if metrics.Metrics[i].MType == "gauge" {
			Value := metrics.Metrics[i].Value
			repo.AddGaugeMetric(metricName, *Value)
		}
	}
	return r
}

func main() {
	ServerConfig.Parse()
	if ServerConfig.StoreInterval != 0 {

		storeTicker := time.NewTicker(ServerConfig.StoreInterval)
		if ServerConfig.Restore {
			RestoreMetricsFromDisk(ServerConfig, Server)
		}

		go func() {
			for {

				<-storeTicker.C
				err := SaveMetricToDisk(ServerConfig, Server)
				if err != nil {
					return
				}

			}
		}()
	}
	r := chi.NewRouter()
	r.Use(middleware.GzipHandle)
	r.Use(middleware.CheckHash(ServerConfig))
	r.Get("/", listMetrics)
	r.Route("/update", func(r chi.Router) {
		r.Post("/", receiveMetricJSON)
		r.Post("/{metricType}/{metricName}/{metricValue}", receiveMetric)
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", valueOfMetricJSON)
		r.Get("/{metricType}/{metricName}", valueOfMetric)
	})
	http.ListenAndServe(ServerConfig.Address, r)
}
