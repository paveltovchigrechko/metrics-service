package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/paveltovchigrechko/metrics-service/internal/config"
	"github.com/paveltovchigrechko/metrics-service/internal/handler"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
)

func main() {
	run()
}

func run() {
	storage := models.NewStorage()
	h := handler.NewHandler(storage)
	r := chi.NewRouter()
	cfg, err := config.SetServerConfig(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	r.Post("/update/{metricsType}/{metricsName}/{metricsValue}", h.PostMetrics)
	r.Get("/", h.MainPage)
	r.Get("/value/{metricsType}/{metricsName}", h.MetricsValue)

	err = http.ListenAndServe(cfg.ServerAddress, r)
	if err != nil {
		log.Fatal(err)
	}
}
