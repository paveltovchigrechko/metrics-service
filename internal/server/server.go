package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/paveltovchigrechko/metrics-service/internal/config"
	"github.com/paveltovchigrechko/metrics-service/internal/handler"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
)

type Server struct {
	cfg *config.ServerConfig
	h   *handler.AppHandler
	r   *chi.Mux
	s   models.Storage
}

func NewServer(c *config.ServerConfig) (*Server, error) {
	storage := models.NewMemStorage()
	if c.Restore {
		err := restoreMetrics(c.FileStoragePath, storage)
		if err != nil && !errors.Is(err, os.ErrNotExist) { // We accept non-existent file on the first start.
			return nil, err
		}
	}

	h := handler.NewHandler(storage)
	r := chi.NewRouter()

	return &Server{
		cfg: c,
		h:   h,
		r:   r,
		s:   storage,
	}, nil
}

func (s *Server) Run() error {
	s.setHandlers()
	return http.ListenAndServe(s.cfg.ServerAddress, s.r)
}

func (s *Server) UseMiddlewares(middlewares ...func(http.Handler) http.Handler) {
	s.r.Use(middlewares...)
}

func (s *Server) setHandlers() {
	s.r.Post("/update/{metricsType}/{metricsName}/{metricsValue}", s.h.PostMetrics)
	s.r.Post("/update", s.h.UpdateEndpoint)
	s.r.Post("/update/", s.h.UpdateEndpoint) // Keep for autotests
	s.r.Post("/value", s.h.ValueEndpoint)
	s.r.Post("/value/", s.h.ValueEndpoint) // Keep for autotests
	s.r.Get("/", s.h.MainPage)
	s.r.Get("/value/{metricsType}/{metricsName}", s.h.MetricsValue)
}

func restoreMetrics(path string, storage models.Storage) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	metrics := make([]models.Metrics, 0)
	err = json.Unmarshal(bytes, &metrics)
	if err != nil {
		return err
	}
	// check for empty slice
	for i := range metrics {
		err := storage.RestoreMetrics(&metrics[i])
		if err != nil {
			return err // We may want to skip a malformed metrics and try to recover the next one.
		}
	}

	return nil
}
