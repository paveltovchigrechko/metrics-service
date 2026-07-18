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

func NewServer(c *config.ServerConfig, middlewares ...func(http.Handler) http.Handler) (*Server, error) {
	storage := models.NewMemStorage()
	if c.Restore {
		err := models.RestoreMetrics(c.FileStoragePath, storage)
		if err != nil && !errors.Is(err, os.ErrNotExist) { // We accept non-existent file on the first start.
			return nil, err
		}
	}

	fs, err := models.NewFileStorage(c.FileStoragePath)
	if err != nil {
		return nil, err
	}

	var h *handler.AppHandler
	if c.StoreInterval == 0 {
		updateFunc := func() error {
			return fs.Save(storage)
		}
		h = handler.NewHandler(storage, updateFunc)
	} else {
		h = handler.NewHandler(storage, nil) // We don't need the callback for synchronous writing to file.
	}

	r := chi.NewRouter()

	s := &Server{
		cfg:         c,
		handler:     h,
		router:      r,
		storage:     storage,
		fileStorage: fs,
	}

	s.useMiddlewares(middlewares...) // Set middlewares before setting handlers
	s.setHandlers()                  // Set handlers

	return s, nil
}

func (s *Server) Run() error {
	if s.cfg.StoreInterval > 0 {
		go s.runStoreLoop() // Separate thread for time ticker
	}

	return http.ListenAndServe(s.cfg.ServerAddress, s.router)
}

func (s *Server) useMiddlewares(middlewares ...func(http.Handler) http.Handler) {
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
