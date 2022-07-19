//TODO:
//на 1 на 1 (назначил на 14 июля 19 00) обсудить реализацию переиспользования подключения к БД
//в целом у меня есть несколько вариантов, но все они мне не нравятся. Реализацию запланирую уже в 4 спринте.

package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/zamibaru89/devops-metrics/internal/config"
	"github.com/zamibaru89/devops-metrics/internal/functions"
	"github.com/zamibaru89/devops-metrics/internal/handlers"
	"github.com/zamibaru89/devops-metrics/internal/middleware"
	"github.com/zamibaru89/devops-metrics/internal/storage"
	"log"
	"net/http"
	"time"
)

func main() {

	ServerConfig, err := config.LoadServerConfig()
	if err != nil {
		log.Fatal(err)
		return
	}
	var Server storage.Repo

	if ServerConfig.DSN != "" {
		Server, err = storage.NewPostgresStorage(ServerConfig)
		if err != nil {
			log.Fatal(err)
			return

		}
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
	r.Route("/updates", func(r chi.Router) {
		r.Post("/", handlers.ReceiveMetricsJSON(ServerConfig, Server))

	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", handlers.ValueOfMetricJSON(ServerConfig, Server))
		r.Get("/{metricType}/{metricName}", handlers.ValueOfMetric(Server))
	})
	http.ListenAndServe(ServerConfig.Address, r)
}
