package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/paveltovchigrechko/metrics-service/internal/handler"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
)

func SetRoutes(r *chi.Mux) {
	storage := models.NewMemStorage()
	h := handler.NewHandler(storage)

	r.Post("/update/{metricsType}/{metricsName}/{metricsValue}", h.PostMetrics)
	r.Post("/update", h.UpdateEndpoint)
	// r.Post("/value", h.ValueEndpoint)
	r.Get("/", h.MainPage)
	r.Get("/value/{metricsType}/{metricsName}", h.MetricsValue)
}
