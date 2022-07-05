package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zamibaru89/devops-metrics/internal/config"
	"github.com/zamibaru89/devops-metrics/internal/functions"
	"github.com/zamibaru89/devops-metrics/internal/storage"

	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"
	"time"
)

var listGauges = []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue"}
var listCounters = []string{"PollCount"}
var AgentConfig = config.AgentConfig{}

var u = &url.URL{
	Scheme: "http",
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

func (m *MetricGauge) SendMetrics(c config.AgentConfig) {

	x := make(map[string]float64)
	j, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err)
		return
	}

	json.Unmarshal(j, &x)
	var metrics storage.MetricStorage

	for _, v := range listGauges {
		value := x[v]
		var Hash string
		if c.Key != "" {
			msg := fmt.Sprintf("%s:gauge:%f", v, value)

			Hash = functions.CreateHash(msg, []byte(c.Key))
		}
		metrics.Metrics = append(metrics.Metrics, storage.Metric{
			ID:    v,
			MType: "gauge",
			Value: &value,
			Hash:  Hash,
		})

	}
	body, _ := json.Marshal(metrics)

	u.Path = path.Join("updates")
	u.Host = c.Address

	sendPOST(*u, body)

}

func (m *MetricCounter) UpdateMetrics() {
	m.PollCount++
}

func (m *MetricCounter) SendMetrics(c config.AgentConfig) {

	xc := make(map[string]int64)
	j, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return
	}
	json.Unmarshal(j, &xc)

	var metrics storage.MetricStorage

	for _, v := range listCounters {
		delta := xc[v]
		var Hash string
		if c.Key != "" {
			msg := fmt.Sprintf("%s:counter:%d", v, delta)

			Hash = functions.CreateHash(msg, []byte(c.Key))
		}
		metrics.Metrics = append(metrics.Metrics, storage.Metric{
			ID:    v,
			MType: "counter",
			Delta: &delta,
			Hash:  Hash,
		})

	}
	body, _ := json.Marshal(metrics)

	u.Path = path.Join("updates")
	u.Host = c.Address

	sendPOST(*u, body)
}

func sendPOST(u url.URL, b []byte) {
	method := "POST"
	client := &http.Client{}

	req, err := http.NewRequest(method, u.String(), bytes.NewBuffer(b))
	req.Header.Add("Content-Type", "application/json")

	if err != nil {
		log.Println(err)
		return
	}

	res, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()

}

func main() {

	var metricG MetricGauge
	var metricC MetricCounter
	AgentConfig.Parse()
	pullTicker := time.NewTicker(AgentConfig.PollInterval)
	pushTicker := time.NewTicker(AgentConfig.ReportInterval)
	sigs := make(chan os.Signal, 4)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	for {
		select {
		case <-pullTicker.C:
			metricG.UpdateMetrics()
			metricC.UpdateMetrics()
			log.Println("running metric.UpdateMetrics()")

		case <-pushTicker.C:

			metricG.SendMetrics(AgentConfig)
			metricC.SendMetrics(AgentConfig)
		case <-sigs:
			log.Println("signal received")
			pullTicker.Stop()
			pushTicker.Stop()
			log.Println("Graceful shutdown")
			os.Exit(1)

		}

	}

}
