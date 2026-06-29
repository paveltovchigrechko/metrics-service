package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

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

	r.Post("/update/{metricsType}/{metricsName}/{metricsValue}", h.PostMetrics)
	r.Get("/", h.MainPage)

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		log.Fatal(err)
	}
}
