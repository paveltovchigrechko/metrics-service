package server

import (
	"net/http"

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

func NewServer(c *config.ServerConfig) *Server {
	storage := models.NewMemStorage()
	h := handler.NewHandler(storage)

	r := chi.NewRouter()
	return &Server{
		cfg: c,
		h:   h,
		r:   r,
		s:   storage,
	}
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
