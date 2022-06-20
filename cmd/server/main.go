package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zamibaru89/devops-metrics/internal/storage"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var Server = storage.NewMemoryStorage()
var Cmd = &cobra.Command{}

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
	} else if metricType == "counter" {

		Value, err := strconv.ParseInt(metricValue, 0, 64)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
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
	metricType := chi.URLParam(r, "metricType")
	if metricType == "gauge" {
		value, err := Server.GetGauge(metricName)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)

		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, value)
	} else if metricType == "counter" {
		value, err := Server.GetCounter(metricName)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusNotFound)

		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, value)
	}

}
func listMetrics(w http.ResponseWriter, r *http.Request) {
	json, _ := json.Marshal(Server.AsJson())
	w.Header().Set("Content-Type", "application/json")
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
	if m.MType == "gauge" {
		if m.Value == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
		} else {
			Server.AddGaugeMetric(m.ID, *m.Value)
			w.Header().Set("Content-Type", "application/json")
			render.JSON(w, r, m)
		}
	} else if m.MType == "counter" {
		if m.Delta == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, m)
		} else {
			Server.AddCounterMetric(m.ID, *m.Delta)
			w.Header().Set("Content-Type", "application/json")
			render.JSON(w, r, m)
		}
	} else {
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

	if m.MType == "counter" {
		value, _ := Server.GetCounter(m.ID)
		m.Delta = &value
		w.Header().Set("Content-Type", "application/json")
		render.JSON(w, r, m)
	} else if m.MType == "gauge" {
		value, _ := Server.GetGauge(m.ID)
		m.Value = &value
		w.Header().Set("Content-Type", "application/json")
		render.JSON(w, r, m)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
	}

}

type Config struct {
	ADDRESS       string `mapstructure:"ADDRESS"`
	FilePath      string `mapstructure:"STORE_FILE "`
	StoreInterval string `mapstructure:"STORE_INTERVAL"`
	Restore       bool   `mapstructure:"RESTORE"`
}

func LoadConfig() (config Config, err error) {
	Cmd.PersistentFlags().StringVarP(&config.ADDRESS, "ADDRESS", "a", "", "URL:PORT")
	Cmd.PersistentFlags().StringVarP(&config.FilePath, "STORE_FILE", "f", "", "Save to filepath?")
	Cmd.PersistentFlags().StringVarP(&config.StoreInterval, "STORE_INTERVAL", "i", "", "Store interval in seconds")
	Cmd.PersistentFlags().BoolVarP(&config.Restore, "RESTORE", "r", true, "Restore from File true/false")

	viper.SetDefault("ADDRESS", ":8080")
	viper.SetDefault("STORE_FILE ", "/tmp/devops-metrics-db.json")
	viper.SetDefault("STORE_INTERVAL", "300s")
	viper.SetDefault("RESTORE", true)

	viper.BindPFlag("ADDRESS", Cmd.PersistentFlags().Lookup("ADDRESS"))
	viper.BindPFlag("STORE_FILE", Cmd.PersistentFlags().Lookup("STORE_FILE"))
	viper.BindPFlag("STORE_INTERVAL", Cmd.PersistentFlags().Lookup("STORE_INTERVAL"))
	viper.BindPFlag("RESTORE", Cmd.PersistentFlags().Lookup("RESTORE"))
	viper.AutomaticEnv()
	Cmd.Execute()
	err = viper.Unmarshal(&config)
	return
}

func SaveMetricToDisk(config Config, m storage.Repo) {

	filePath := config.FilePath
	fileBits := os.O_WRONLY | os.O_CREATE | os.O_TRUNC

	file, err := os.OpenFile(filePath, fileBits, 0600)
	if err != nil {
		log.Fatal(err)
	}

	data, err := json.Marshal(m.AsJson())
	if err != nil {
		log.Fatal(err)
	}

	_, err = file.Write(data)
	if err != nil {
		log.Fatal(err)
	}

	file.Close()

}

func RestoreMetricsFromDisk(config Config, r storage.Repo) storage.Repo {
	repo := r
	path := config.FilePath

	data, _ := ioutil.ReadFile(path)

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
	config, _ := LoadConfig()
	storeDuration, _ := time.ParseDuration(config.StoreInterval)
	storeTicker := time.NewTicker(storeDuration)
	if config.Restore == true {
		RestoreMetricsFromDisk(config, Server)
	}
	go func() {
		for {
			select {
			case <-storeTicker.C:
				SaveMetricToDisk(config, Server)
				fmt.Println("Save to disk")
			}
		}
	}()
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
