package server

import (
	"database/sql"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/paveltovchigrechko/metrics-service/internal/config"
	"github.com/paveltovchigrechko/metrics-service/internal/config/db"
	"github.com/paveltovchigrechko/metrics-service/internal/handler"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
)

type storageMode int

const (
	postgresMode storageMode = iota
	fileMode
	memoryMode
)

type Server struct {
	cfg         *config.ServerConfig
	handler     *handler.AppHandler
	router      *chi.Mux
	storage     models.Storage
	fileStorage *models.FileStorage
	closer      io.Closer // This is for closing database
}

func New(c *config.ServerConfig, middlewares ...func(http.Handler) http.Handler) (*Server, error) {
	mode := defineMode(c)

	// Set up storage
	var database *sql.DB
	var storage models.Storage
	var fs *models.FileStorage
	var updateFunc func() error
	var err error
	var closer io.Closer

	switch mode {
	case postgresMode:
		database, err = db.OpenDB(c.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		err = db.RunMigrations(database)
		if err != nil {
			database.Close()
			return nil, err
		}
		storage = models.NewPostgresStorage(database)
		closer = database
	case fileMode:
		storage = models.NewMemStorage()
		fs, err = models.NewFileStorage(c.FileStoragePath)
		if err != nil {
			return nil, err
		}

		if c.Restore {
			err = fs.Load(storage)
			if err != nil && !errors.Is(err, os.ErrNotExist) { // We accept that file doesn't exist on first startup.
				return nil, err
			}
		}

		if c.StoreInterval == 0 {
			updateFunc = func() error {
				return fs.Save(storage)
			}
		}
	case memoryMode:
		storage = models.NewMemStorage()
	}

	var pinger handler.Pinger
	if database != nil { //Should I check here for type == PostgresStorage and create pinger from ps.db?
		pinger = database
	}

	h := handler.NewHandler(storage, pinger, updateFunc)

	r := chi.NewRouter()

	s := &Server{
		cfg:         c,
		handler:     h,
		router:      r,
		storage:     storage,
		fileStorage: fs,
		closer:      closer,
	}

	s.useMiddlewares(middlewares...) // Set middlewares before setting handlers
	s.setHandlers()                  // Set handlers

	return s, nil
}

func (s *Server) Run() error {
	if s.cfg.StoreInterval > 0 && s.fileStorage != nil {
		go s.runStoreLoop() // Separate thread for time ticker
	}

	return http.ListenAndServe(s.cfg.ServerAddress, s.router)
}

func (s *Server) StopDB() error {
	if s.closer == nil { // Use storage(type).PostgresStorage.db
		return nil
	}
	return s.closer.Close()
}

func (s *Server) useMiddlewares(middlewares ...func(http.Handler) http.Handler) {
	s.router.Use(middlewares...)
}

func (s *Server) setHandlers() {
	s.router.Post("/update/{metricsType}/{metricsName}/{metricsValue}", s.handler.PostMetrics)
	s.router.Post("/update", s.handler.UpdateEndpoint)
	s.router.Post("/update/", s.handler.UpdateEndpoint) // Keep for autotests
	s.router.Post("/updates", s.handler.UpdatesEndpoint)
	s.router.Post("/value", s.handler.ValueEndpoint)
	s.router.Post("/value/", s.handler.ValueEndpoint) // Keep for autotests
	s.router.Get("/", s.handler.MainPage)
	s.router.Get("/value/{metricsType}/{metricsName}", s.handler.MetricsValue)
	s.router.Get("/ping", s.handler.PingEndpoint)
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

func defineMode(c *config.ServerConfig) storageMode {
	if c.DatabaseDSN != "" {
		return postgresMode
	}

	if c.FileStoragePath != "" {
		return fileMode
	}

	return memoryMode
}
