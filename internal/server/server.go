package server

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/paveltovchigrechko/metrics-service/internal/config"
	"github.com/paveltovchigrechko/metrics-service/internal/handler"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
)

type Server struct {
	cfg         *config.ServerConfig
	handler     *handler.AppHandler
	router      *chi.Mux
	storage     models.Storage
	fileStorage *models.FileStorage
}

func NewServer(c *config.ServerConfig) (*Server, error) {
	storage := models.NewMemStorage()
	if c.Restore {
		err := models.RestoreMetrics(c.FileStoragePath, storage)
		if err != nil && !errors.Is(err, os.ErrNotExist) { // We accept non-existent file on the first start.
			return nil, err
		}
	}

	var h *handler.AppHandler
	fs := &models.FileStorage{Path: c.FileStoragePath}
	if c.StoreInterval == 0 {
		updateFunc := func() error {
			return fs.Save(storage)
		}
		h = handler.NewHandler(storage, updateFunc)
	} else {
		h = handler.NewHandler(storage, nil)
	}

	r := chi.NewRouter()

	return &Server{
		cfg:         c,
		handler:     h,
		router:      r,
		storage:     storage,
		fileStorage: fs,
	}, nil
}

func (s *Server) Run() error {
	s.setHandlers()

	if s.cfg.StoreInterval > 0 {
		go s.runStoreLoop() // Separate thread for time ticker
	}

	return http.ListenAndServe(s.cfg.ServerAddress, s.router)
}

func (s *Server) UseMiddlewares(middlewares ...func(http.Handler) http.Handler) {
	s.router.Use(middlewares...)
}

func (s *Server) setHandlers() {
	s.router.Post("/update/{metricsType}/{metricsName}/{metricsValue}", s.handler.PostMetrics)
	s.router.Post("/update", s.handler.UpdateEndpoint)
	s.router.Post("/update/", s.handler.UpdateEndpoint) // Keep for autotests
	s.router.Post("/value", s.handler.ValueEndpoint)
	s.router.Post("/value/", s.handler.ValueEndpoint) // Keep for autotests
	s.router.Get("/", s.handler.MainPage)
	s.router.Get("/value/{metricsType}/{metricsName}", s.handler.MetricsValue)
}

func (s *Server) runStoreLoop() {
	ticker := time.NewTicker(s.cfg.StoreInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.fileStorage.Save(s.storage); err != nil {
			log.Printf("save metrics error: %v", err)
		}
	}
}
