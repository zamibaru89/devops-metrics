package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

var GaugeMemory map[string]float64
var CounterMemory map[string]int64

func receiveMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	if metricType == "gauge" {
		var receivedMetric MetricsGauge
		var err error
		receivedMetric.ID = metricName
		receivedMetric.Value, err = strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		GaugeMemory[receivedMetric.ID] = receivedMetric.Value

	} else if metricType == "counter" {
		var receivedMetric MetricsCounter
		receivedMetric.ID = metricName
		var err error
		receivedMetric.Value, err = strconv.ParseInt(metricValue, 0, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		previousValue := CounterMemory[receivedMetric.ID]
		CounterMemory[receivedMetric.ID] = receivedMetric.Value + previousValue

	} else {
		w.WriteHeader(501)
	}

}

func valueOfMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	if metricType == "counter" {
		if value, ok := CounterMemory[metricName]; ok {
			fmt.Fprintln(w, value)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else if metricType == "gauge" {
		if value, ok := GaugeMemory[metricName]; ok {
			fmt.Fprintln(w, value)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)

	}
}
func listMetrics(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintln(w, "#########GAUGE METRICS#########")
	for key, value := range GaugeMemory {
		fmt.Fprintln(w, key, value)

	}
	fmt.Fprintln(w, "#########COUNTER METRICS#########")
	for key, value := range CounterMemory {
		fmt.Fprintln(w, key, value)

	}

}

type MetricsGauge struct {
	ID    string
	Value float64
}
type MetricsCounter struct {
	ID    string
	Value int64
}

func main() {
	GaugeMemory = make(map[string]float64)
	CounterMemory = make(map[string]int64)
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", listMetrics)
		r.Post("/{operation}/", func(w http.ResponseWriter, r *http.Request) {
			operation := chi.URLParam(r, "operation")
			//fmt.Println(operation)
			if operation != "update" {
				w.WriteHeader(404)
			} else if operation != "value" {
				w.WriteHeader(404)
			}

		})
		r.Post("/update/{metricType}/*", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println()
			w.WriteHeader(404)

		})
		r.Post("/update/{metricType}/{metricName}/{metricValue}", receiveMetric)
		r.Get("/value/{metricType}/{metricName}", valueOfMetric)
	})

	http.ListenAndServe(":8080", r)
}

//Сервер должен возвращать текущее значение запрашиваемой метрики в текстовом виде по запросу
//GET http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ> (со статусом http.StatusOK).
//При попытке запроса неизвестной серверу метрики сервер должен возвращать http.StatusNotFound.
//По запросу GET http://<АДРЕС_СЕРВЕРА>/ сервер должен отдавать HTML-страничку со списком имён и значений всех известных ему на текущий момент метрик.
