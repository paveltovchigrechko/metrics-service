package router

import (
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/paveltovchigrechko/metrics-service/internal/config"
	"github.com/paveltovchigrechko/metrics-service/internal/handler"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
)

func PrepareServerRouterAndConfig() (*chi.Mux, *config.ServerConfig, error) {
	cfg, err := config.SetServerConfig(os.Args[1:])
	if err != nil {
		return nil, nil, err
	}

	storage := models.NewStorage()
	h := handler.NewHandler(storage)
	r := chi.NewRouter()

	r.Post("/update/{metricsType}/{metricsName}/{metricsValue}", h.PostMetrics)
	r.Get("/", h.MainPage)
	r.Get("/value/{metricsType}/{metricsName}", h.MetricsValue)

	return r, cfg, nil
}
