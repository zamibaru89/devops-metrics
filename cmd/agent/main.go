package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var list = []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "PollCount", "RandomValue"}

type Metric struct {
	runtime.MemStats
	PollCount   uint64
	RandomValue uint64
}

func (m *Metric) UpdateMetrics() {
	runtime.ReadMemStats(&m.MemStats)
	m.PollCount++
	m.RandomValue = uint64(rand.Intn(1000))
}

func (m *Metric) SendMetrics() {
	x := make(map[string]float64)
	j, _ := json.Marshal(m)
	json.Unmarshal(j, &x)

	for _, v := range list {
		value := x[v]
		fmt.Println("key: ", v, "value: ", value)

	}

}

func main() {
	var metric Metric

	pullTicker := time.NewTicker(2 * time.Second)
	pushTicker := time.NewTicker(10 * time.Second)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	for {
		select {
		case <-pullTicker.C:
			metric.UpdateMetrics()
			fmt.Println("running metric.UpdateMetrics()")

		case <-pushTicker.C:
			metric.SendMetrics()

		case <-sigs:
			fmt.Println("signal received")
			pullTicker.Stop()
			pushTicker.Stop()
			fmt.Println("Graceful shutdown")
			os.Exit(1)

		}

	}

}
