//TODO:
//Обсудить на 1 на 1 как реализовать
//1) Фикс дупликацию с хендлером и вынос бизнес логики
//2) Реализация апи клиента
//3) config.GetAgentConfig() и варианты реализации

package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/zamibaru89/devops-metrics/internal/config"
	"github.com/zamibaru89/devops-metrics/internal/functions"
	"github.com/zamibaru89/devops-metrics/internal/handlers"
	"github.com/zamibaru89/devops-metrics/internal/middleware"
	"github.com/zamibaru89/devops-metrics/internal/storage"
	"net/http"
	"time"
)

func main() {

	ServerConfig, _ := config.LoadServerConfig()
	var Server storage.Repo
	if ServerConfig.DSN != "" {
		Server = storage.NewPostgresStorage(ServerConfig)
	} else {
		Server = storage.NewMemoryStorage()
	}
	if ServerConfig.Restore && ServerConfig.DSN == "" {
		functions.RestoreMetricsFromDisk(ServerConfig, Server)
	}
	if ServerConfig.StoreInterval != 0 && ServerConfig.DSN == "" {

		storeTicker := time.NewTicker(ServerConfig.StoreInterval)

		go func() {
			for {

				<-storeTicker.C
				err := functions.SaveMetricToDisk(ServerConfig, Server)
				if err != nil {
					return
				}

			}
		}()
	}
	r := chi.NewRouter()
	r.Use(middleware.GzipHandle)
	//r.Use(middleware.CheckHash(ServerConfig))
	r.Get("/", handlers.ListMetrics(Server))
	r.Get("/ping", handlers.PingDB(ServerConfig))

	r.With(middleware.CheckHash(ServerConfig)).Route("/update", func(r chi.Router) {
		r.Post("/", handlers.ReceiveMetricJSON(ServerConfig, Server))
		r.Post("/{metricType}/{metricName}/{metricValue}", handlers.ReceiveMetric(ServerConfig, Server))
	})
	r.With(middleware.CheckHash(ServerConfig)).Route("/updates", func(r chi.Router) {
		r.Post("/", handlers.ReceiveMetricsJSON(ServerConfig, Server))

	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", handlers.ValueOfMetricJSON(ServerConfig, Server))
		r.Get("/{metricType}/{metricName}", handlers.ValueOfMetric(Server))
	})
	http.ListenAndServe(ServerConfig.Address, r)
}
