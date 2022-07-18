package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
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
type CPU struct {
	CPU     string
	percent float64
}

type CPUs struct {
	CPUS []CPU
}

type MetricGauge struct {
	runtime.MemStats
	CPUs
	TotalMemory float64
	FreeMemory  float64
	RandomValue float64
}

func (m *MetricGauge) UpdateMetrics() {
	runtime.ReadMemStats(&m.MemStats)
	m.RandomValue = rand.Float64()

}

func (m *MetricGauge) UpdateMetricsPSUtils() error {
	cpu, _ := cpu.Percent(0, true)

	for index, value := range cpu {
		m.CPUs.CPUS = append(m.CPUs.CPUS, CPU{
			CPU:     fmt.Sprintf("CPUutilization%d", index+1),
			percent: value,
		})
	}
	memoryStat, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	m.TotalMemory = float64(memoryStat.Total)
	m.FreeMemory = float64(memoryStat.Free)
	return nil

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

	body, err := json.Marshal(metrics.Metrics)
	if err != nil {
		log.Println(err)
		return
	}
	u.Path = path.Join("updates")
	u.Host = c.Address

	sendPOST(*u, body)

}

func (m *MetricGauge) SendMetricsPSUtils(c config.AgentConfig) {
	var metrics storage.MetricStorage
	for _, value := range m.CPUs.CPUS {
		var Hash string
		if c.Key != "" {
			msg := fmt.Sprintf("%s:gauge:%f", value.CPU, value.percent)

			Hash = functions.CreateHash(msg, []byte(c.Key))
		}
		metrics.Metrics = append(metrics.Metrics, storage.Metric{
			ID:    "FreeMemory",
			MType: "gauge",
			Value: &m.FreeMemory,
			Hash:  Hash,
		})
		metrics.Metrics = append(metrics.Metrics, storage.Metric{
			ID:    "TotalMemory",
			MType: "gauge",
			Value: &m.TotalMemory,
			Hash:  Hash,
		})
	}

	body, err := json.Marshal(metrics.Metrics)
	if err != nil {
		log.Println(err)
		return
	}
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

	body, err := json.Marshal(metrics.Metrics)
	if err != nil {
		log.Println(err)
		return
	}
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
			//metricG.UpdateMetrics()
			metricG.UpdateMetricsPSUtils()
			metricC.UpdateMetrics()
			log.Println("running metric.UpdateMetrics()")

		case <-pushTicker.C:

			//metricG.SendMetrics(AgentConfig)
			//metricC.SendMetrics(AgentConfig)
			metricG.SendMetricsPSUtils(AgentConfig)
		case <-sigs:
			log.Println("signal received")
			pullTicker.Stop()
			pushTicker.Stop()
			log.Println("Graceful shutdown")
			os.Exit(1)

		}

	}

}
