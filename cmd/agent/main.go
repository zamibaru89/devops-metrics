package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

var Cmd = &cobra.Command{}
var listGauges = []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue"}
var listCounters = []string{"PollCount"}

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

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m *MetricGauge) SendMetrics(c Config) {

	x := make(map[string]float64)
	j, _ := json.Marshal(m)

	json.Unmarshal(j, &x)

	for _, v := range listGauges {
		value := x[v]
		var m Metrics
		m.ID = v
		m.MType = "gauge"
		m.Value = &value
		body, _ := json.Marshal(m)
		//u.Path = path.Join("update", "gauge", v, fmt.Sprintf("%f", value))
		u.Path = path.Join("update")
		u.Host = c.ADDRESS
		sendPOST(*u, body)
	}

}

func (m *MetricCounter) UpdateMetrics() {
	m.PollCount++
}

func (m *MetricCounter) SendMetrics(c Config) {

	xc := make(map[string]int64)
	j, _ := json.Marshal(m)
	json.Unmarshal(j, &xc)

	for _, v := range listCounters {
		delta := xc[v]
		var m Metrics
		m.ID = v
		m.MType = "counter"
		m.Delta = &delta
		body, _ := json.Marshal(m)
		//u.Path = path.Join("update", "counter", v, strconv.FormatInt(value, 10))
		u.Path = path.Join("update")
		u.Host = c.ADDRESS
		sendPOST(*u, body)
	}

}

func sendPOST(u url.URL, b []byte) {
	method := "POST"
	client := &http.Client{}

	req, err := http.NewRequest(method, u.String(), bytes.NewBuffer(b))
	req.Header.Add("Content-Type", "application/json")

	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := client.Do(req)
	fmt.Println(req, res)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

}

type Config struct {
	ADDRESS        string `mapstructure:"ADDRESS"`
	ReportInterval string `mapstructure:"REPORT_INTERVAL"`
	PollInterval   string `mapstructure:"POLL_INTERVAL"`
}

func LoadConfig() (config Config, err error) {

	Cmd.PersistentFlags().StringVarP(&config.ReportInterval, "REPORT_INTERVAL", "r", "", "10s")
	Cmd.PersistentFlags().StringVarP(&config.ADDRESS, "ADDRESS", "a", "", "URL:PORT")
	Cmd.PersistentFlags().StringVarP(&config.PollInterval, "POLL_INTERVAL", "p", "", "2s")

	viper.SetDefault("REPORT_INTERVAL", "10s")
	viper.SetDefault("ADDRESS", "localhost:8080")
	viper.SetDefault("POLL_INTERVAL", "2s")

	viper.BindPFlag("REPORT_INTERVAL", Cmd.PersistentFlags().Lookup("REPORT_INTERVAL"))
	viper.BindPFlag("ADDRESS", Cmd.PersistentFlags().Lookup("ADDRESS"))
	viper.BindPFlag("POLL_INTERVAL", Cmd.PersistentFlags().Lookup("POLL_INTERVAL"))
	viper.AutomaticEnv()
	Cmd.Execute()
	err = viper.Unmarshal(&config)
	return
}

func main() {

	var metricG MetricGauge
	var metricC MetricCounter
	config, _ := LoadConfig()

	pullDuration, _ := time.ParseDuration(config.PollInterval)
	reportDuration, _ := time.ParseDuration(config.ReportInterval)
	pullTicker := time.NewTicker(pullDuration)
	pushTicker := time.NewTicker(reportDuration)
	sigs := make(chan os.Signal, 4)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	for {
		select {
		case <-pullTicker.C:
			metricG.UpdateMetrics()
			metricC.UpdateMetrics()
			fmt.Println("running metric.UpdateMetrics()")

		case <-pushTicker.C:
			metricG.SendMetrics(config)
			metricC.SendMetrics(config)
		case <-sigs:
			fmt.Println("signal received")
			pullTicker.Stop()
			pushTicker.Stop()
			fmt.Println("Graceful shutdown")
			os.Exit(1)

		}

	}

}
