package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/spf13/viper"
	"github.com/zamibaru89/devops-metrics/internal/storage"
	"net/http"
	"strconv"
)

var Server = storage.NewMemoryStorage()

func receiveMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	if metricType == "gauge" {
		Value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		Server.AddGaugeMetric(metricName, Value)
	} else if metricType == "counter" {

		Value, err := strconv.ParseInt(metricValue, 0, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		Server.AddCounterMetric(metricName, Value)

	} else {
		w.WriteHeader(501)
	}

}

func valueOfMetric(w http.ResponseWriter, r *http.Request) {

	metricName := chi.URLParam(r, "metricName")
	metricType := chi.URLParam(r, "metricName")
	if metricType == "gauge" {
		value, err := Server.GetGauge(metricName)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)

		}
		fmt.Fprintln(w, value)
	} else if metricType == "counter" {
		value, err := Server.GetCounter(metricName)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)

		}
		fmt.Fprintln(w, value)
	}

}
func listMetrics(w http.ResponseWriter, r *http.Request) {
	json, _ := json.Marshal(Server.AsJson())
	fmt.Fprintln(w, string(json))
}

func receiveMetricJSON(w http.ResponseWriter, r *http.Request) {
	var m storage.Metric
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {

		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, m)
		return
	}
	if m.MType == "gauge" {
		if m.Value == nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			Server.AddGaugeMetric(m.ID, *m.Value)
			render.JSON(w, r, m)
		}
	} else if m.MType == "counter" {
		if m.Delta == nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, m)
		} else {
			Server.AddCounterMetric(m.ID, *m.Delta)
			render.JSON(w, r, m)
		}
	} else {
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

	if m.MType == "counter" {
		value, _ := Server.GetCounter(m.ID)
		m.Delta = &value
		render.JSON(w, r, m)
	} else if m.MType == "gauge" {
		value, _ := Server.GetGauge(m.ID)
		m.Value = &value
		render.JSON(w, r, m)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

}

type Config struct {
	ADDRESS string `mapstructure:"ADDRESS"`
}

func LoadConfig() (config Config, err error) {
	viper.SetDefault("ADDRESS", ":8080")
	viper.AutomaticEnv()

	err = viper.Unmarshal(&config)
	return
}

func main() {
	config, _ := LoadConfig()

	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Get("/", listMetrics)
	r.Route("/update", func(r chi.Router) {
		r.Post("/", receiveMetricJSON)
		r.Post("/{metricType}/{metricName}/{metricValue}", receiveMetric)
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", valueOfMetricJSON)
		r.Get("/{metricType}/{metricName}", valueOfMetric)
	})
	http.ListenAndServe(config.ADDRESS, r)
}
