package middleware

import (
	"bytes"
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"github.com/zamibaru89/devops-metrics/internal/config"
	"github.com/zamibaru89/devops-metrics/internal/functions"
	"github.com/zamibaru89/devops-metrics/internal/storage"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func CheckHash(config config.ServerConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Key != "" {
				var metric storage.Metric
				var msg string
				data, err := io.ReadAll(r.Body)
				r.Body = ioutil.NopCloser(bytes.NewReader(data))
				if err != nil {
					log.Println(err)
					return
				}
				err = json.Unmarshal(data, &metric)
				if err != nil {
					log.Println(err)
					return
				}
				if metric.MType == "counter" {
					counter := *metric.Delta
					msg = fmt.Sprintf("%s:counter:%d", metric.ID, counter)
				} else if metric.MType == "gauge" {
					gauge := *metric.Value
					msg = fmt.Sprintf("%s:gauge:%f", metric.ID, gauge)
				}
				hash := functions.CreateHash(msg, []byte(config.Key))
				if !hmac.Equal([]byte(hash), []byte(metric.Hash)) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
			next.ServeHTTP(w, r)

		})
	}
}
