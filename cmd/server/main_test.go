package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_receiveMetric(t *testing.T) {
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
			r.Post("/update/{metricType}/{metricName}/{metricValue}", receiveMetric)
			ts := httptest.NewServer(r)
			defer ts.Close()
			// проверяем код ответа
			testLink := ts.URL + tt.want.path

			req, _ := http.NewRequest(http.MethodPost, testLink, nil)
			req.Header.Set("Content-Type", "Content-Type: text/plain")
			res, _ := http.DefaultClient.Do(req)

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, res.StatusCode)

			}
			res.Body.Close()
		})
	}
}
func Test_listMetrics(t *testing.T) {
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
			r.Get("/", listMetrics)
			r.Post("/update/{metricType}/{metricName}/{metricValue}", receiveMetric)
			ts := httptest.NewServer(r)
			defer ts.Close()
			// проверяем код ответа
			testLink := ts.URL + tt.want.path

			req, _ := http.NewRequest(http.MethodGet, testLink, nil)
			req.Header.Set("Content-Type", "Content-Type: text/plain")
			res, _ := http.DefaultClient.Do(req)

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, res.StatusCode)

			}
			res.Body.Close()
		})
	}
}
