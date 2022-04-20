package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

var GaugeMemory map[string]float64
var CounterMemory map[string]int64

func receiveGauge(w http.ResponseWriter, r *http.Request) {
	//s := "/update/gauge/alloc/12"
	url := r.URL.Path
	var receivedMetric MetricsGauge
	parsedUrl := strings.Split(url, "/")
	receivedMetric.ID = parsedUrl[4]
	receivedMetric.ID = parsedUrl[3]

	receivedMetric.Value, _ = strconv.ParseFloat(parsedUrl[4], 64)
	if receivedMetric.ID == " " {
		w.WriteHeader(404)
	}

	fmt.Printf("%+v\n", receivedMetric)
	GaugeMemory[receivedMetric.ID] = receivedMetric.Value

	for _, value := range GaugeMemory {
		fmt.Println(value)
	}

}

func receiveCounter(w http.ResponseWriter, r *http.Request) {
	//s := "/update/gauge/alloc/12"
	url := r.URL.Path
	var receivedMetric MetricsCounter
	parsedUrlCounter := strings.Split(url, "/")
	receivedMetric.ID = parsedUrlCounter[4]
	receivedMetric.ID = parsedUrlCounter[3]

	receivedMetric.Value, _ = strconv.ParseInt(parsedUrlCounter[4], 0, 64)
	_, err := strconv.ParseInt(parsedUrlCounter[4], 0, 64)
	if err != nil {
		fmt.Println("ËRROR IS", err)
	}
	fmt.Println("Parsed URL is:", parsedUrlCounter)
	fmt.Printf("%+v\n", receivedMetric)
	previousValue := CounterMemory[receivedMetric.ID]
	fmt.Println("Previous was ", previousValue)
	CounterMemory[receivedMetric.ID] = receivedMetric.Value + previousValue

	for _, value := range CounterMemory {
		fmt.Println(value)
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
	http.HandleFunc("/update/gauge/", receiveGauge)
	http.HandleFunc("/update/counter/", receiveCounter)
	http.HandleFunc("/metrics/", listMetrics)
	http.ListenAndServe(":8080", nil)
}

//Сервер должен собирать и хранить произвольные метрики двух типов:
//gauge, тип float64, новое значение должно замещать предыдущее;
//counter, тип int64, новое значение должно добавляться к предыдущему (если оно ранее уже было известно серверу).
//Метрики должны приниматься сервером по протоколу HTTP, методом POST:
//по умолчанию открывать порт 8080 на адресе 127.0.0.1;
//в формате http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>;
//Content-Type: text/plain;
//при успешном приёме возвращать статус: http.StatusOK.
