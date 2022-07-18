package functions

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/zamibaru89/devops-metrics/internal/config"
	"github.com/zamibaru89/devops-metrics/internal/storage"
	"io/ioutil"
	"log"
	"os"
)

func CreateHash(message string, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	hash := hex.EncodeToString(h.Sum(nil))
	return hash
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
