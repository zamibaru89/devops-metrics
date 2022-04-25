package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

var listGauges = []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue"}
var listCounters = []string{"PollCount"}

var serverToSendAddress string = "127.0.0.1:8080"

var u = &url.URL{
	Scheme: "http",
	Host:   "localhost:8080",
}

type MetricCounter struct {
	PollCount uint64
}

type MetricGauge struct {
	runtime.MemStats
	RandomValue float64
}

func (m *MetricGauge) UpdateMetrics() {
	runtime.ReadMemStats(&m.MemStats)
	m.RandomValue = rand.Float64()
}

func (m *MetricGauge) SendMetrics() {
	x := make(map[string]interface{})
	j, _ := json.Marshal(m)
	json.Unmarshal(j, &x)

	for _, v := range listGauges {
		value := x[v]

		u.Path = path.Join("update", "gauge", v, fmt.Sprintf("%f", value))
		sendPOST(*u)
	}

}

func (m *MetricCounter) UpdateMetrics() {
	m.PollCount++
}

func (m *MetricCounter) SendMetrics() {
	xc := make(map[string]int64)
	j, _ := json.Marshal(m)
	json.Unmarshal(j, &xc)

	for _, v := range listCounters {
		value := xc[v]

		u.Path = path.Join("update", "counter", v, strconv.FormatInt(value, 10))
		sendPOST(*u)
	}

}

func sendPOST(u url.URL) {
	method := "POST"
	client := &http.Client{}

	req, err := http.NewRequest(method, u.String(), nil)
	req.Header.Add("Content-Type", "Content-Type: text/plain")
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

}
func main() {
	var metricG MetricGauge
	var metricC MetricCounter

	pullTicker := time.NewTicker(2 * time.Second)
	pushTicker := time.NewTicker(10 * time.Second)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	for {
		select {
		case <-pullTicker.C:
			metricG.UpdateMetrics()
			metricC.UpdateMetrics()
			fmt.Println("running metric.UpdateMetrics()")

		case <-pushTicker.C:
			metricG.SendMetrics()
			metricC.SendMetrics()
		case <-sigs:
			fmt.Println("signal received")
			pullTicker.Stop()
			pushTicker.Stop()
			fmt.Println("Graceful shutdown")
			os.Exit(1)

		}

	}

}
