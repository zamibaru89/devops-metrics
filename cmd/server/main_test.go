package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/zamibaru89/devops-metrics/internal/config"
	"github.com/zamibaru89/devops-metrics/internal/handlers"
	"github.com/zamibaru89/devops-metrics/internal/storage"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_receiveMetric(t *testing.T) {
	var ServerConfig config.ServerConfig
	var Server storage.Repo = storage.NewMemoryStorage()

	type want struct {
		code int
		path string
	}

	tests := []struct {
		name string
		want want
	}{

		{
			name: "Positive test update of gauge",
			want: want{
				code: 200,
				path: "/update/gauge/alloc/1",
			},
		},
		{
			name: "Positive test update of counter",
			want: want{
				code: 200,
				path: "/update/counter/somemetric/100",
			},
		},
		{
			name: "Invalid metric type",
			want: want{
				code: 501,
				path: "/update/metric/somemetric2/100.10",
			},
		},
		{
			name: "Invalid operation",
			want: want{
				code: 404,
				path: "/delete",
			},
		},
		{
			name: "Without ID",
			want: want{
				code: 404,
				path: "/update/gauge/somemetric2/",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := chi.NewRouter()
			r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.ReceiveMetric(ServerConfig, Server))
			ts := httptest.NewServer(r)
			defer ts.Close()
			// проверяем код ответа
			testLink := ts.URL + tt.want.path

			req, err := http.NewRequest(http.MethodPost, testLink, nil)
			if err != nil {
				log.Println(err)
				return
			}
			req.Header.Set("Content-Type", "Content-Type: text/plain")
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Println(err)
				return
			}
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, res.StatusCode)

			}
			res.Body.Close()
		})
	}
}

func Test_receiveMetricJSON(t *testing.T) {
	var ServerConfig config.ServerConfig
	var Server storage.Repo = storage.NewMemoryStorage()
	type want struct {
		code int
		path string
		body string
	}

	tests := []struct {
		name string
		want want
	}{

		{
			name: "Positive test update of gauge",
			want: want{
				code: 200,
				path: "/update",
				body: "{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":2.0}",
			},
		},
		{
			name: "Positive test update of counter",
			want: want{
				code: 200,
				path: "/update",
				body: "{\"id\":\"PollCount\",\"type\":\"counter\",\"delta\":1}",
			},
		},
		{
			name: "Invalid metric type",
			want: want{
				code: 400,
				path: "/update",
				body: "{\"id\":\"PollCount\",\"type\":123,\"delta\":1}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := chi.NewRouter()

			r.Post("/update", handlers.ReceiveMetricJSON(ServerConfig, Server))

			ts := httptest.NewServer(r)
			testLink := ts.URL + tt.want.path
			defer ts.Close()

			payload := strings.NewReader(tt.want.body)

			req, err := http.NewRequest(http.MethodPost, testLink, payload)

			if err != nil {
				log.Println(err)
				return
			}
			req.Header.Set("Content-Type", "Content-Type: application/json")
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Println(err)
				return
			}
			if res.StatusCode != tt.want.code {

				t.Errorf("Expected status code %d, got %d", tt.want.code, res.StatusCode)

			}
			res.Body.Close()
		})
	}
}

func Test_listMetrics(t *testing.T) {
	var ServerConfig config.ServerConfig
	var Server storage.Repo = storage.NewMemoryStorage()
	type want struct {
		code int
		path string
	}

	tests := []struct {
		name string
		want want
	}{

		{
			name: "Positive test list of metrics",
			want: want{
				code: 200,
				path: "/",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := chi.NewRouter()
			r.Get("/", handlers.ListMetrics(Server))
			r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.ReceiveMetric(ServerConfig, Server))
			ts := httptest.NewServer(r)
			defer ts.Close()
			// проверяем код ответа
			testLink := ts.URL + tt.want.path

			req, err := http.NewRequest(http.MethodGet, testLink, nil)
			if err != nil {
				log.Println(err)
				return
			}
			req.Header.Set("Content-Type", "Content-Type: text/plain")
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Println(err)
				return
			}

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, res.StatusCode)

			}
			res.Body.Close()
		})
	}
}
