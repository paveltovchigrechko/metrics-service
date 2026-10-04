package server

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/paveltovchigrechko/metrics-service/internal/audit"
	"github.com/paveltovchigrechko/metrics-service/internal/config"
	"github.com/paveltovchigrechko/metrics-service/internal/config/db"
	"github.com/paveltovchigrechko/metrics-service/internal/handler"
	"github.com/paveltovchigrechko/metrics-service/internal/repository"
)

const (
	auditQueueSize  = 1024
	shutdownTimeout = 5 * time.Second
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
	storage     repository.Storage
	fileStorage *repository.FileStorage
	closer      io.Closer // This is for closing database
	auditCloser io.Closer // Flushes queued audit events; nil if audit is disabled.
}

func New(c *config.ServerConfig, middlewares ...func(http.Handler) http.Handler) (*Server, error) {
	mode := defineMode(c)

	// Set up storage
	var database *sql.DB
	var storage repository.Storage
	var fs *repository.FileStorage
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
		storage = repository.NewPostgresStorage(database)
		closer = database
	case fileMode:
		storage = repository.NewMemStorage()
		fs, err = repository.NewFileStorage(c.FileStoragePath)
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
		storage = repository.NewMemStorage()
	}

	var pinger handler.Pinger
	if database != nil {
		pinger = database
	}

	notifier, auditCloser := newAuditNotifier(c)
	h := handler.NewHandler(storage, pinger, updateFunc, notifier)

	r := chi.NewRouter()

	s := &Server{
		cfg:         c,
		handler:     h,
		router:      r,
		storage:     storage,
		fileStorage: fs,
		closer:      closer,
		auditCloser: auditCloser,
	}

	s.useMiddlewares(middlewares...) // Set middlewares before setting handlers
	s.setHandlers()                  // Set handlers

	return s, nil
}

// Run serves HTTP requests until SIGINT/SIGTERM, then shuts down gracefully:
// in-flight requests are finished and their audit events are flushed by StopAudit.
func (s *Server) Run() error {
	if s.cfg.StoreInterval > 0 && s.fileStorage != nil {
		go s.runStoreLoop() // Separate thread for time ticker
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: s.cfg.ServerAddress, Handler: s.router}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}

// StopAudit waits until all queued audit events are delivered.
func (s *Server) StopAudit() error {
	if s.auditCloser == nil {
		return nil
	}

	return s.auditCloser.Close()
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
	s.router.Post("/updates/", s.handler.UpdatesEndpoint) // Keep for autotests
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

// newAuditNotifier builds the audit publisher from the config.
// Every receiver is wrapped in an AsyncObserver, so audit never blocks request handling.
// It returns nils (audit disabled) if neither the audit file nor the audit URL is set.
// The returned io.Closer must be closed on shutdown to flush the queued events.
func newAuditNotifier(c *config.ServerConfig) (audit.Notifier, io.Closer) {
	var observers []audit.Observer
	if c.AuditFilePath != "" {
		observers = append(observers, audit.NewAsyncObserver(audit.NewFileObserver(c.AuditFilePath), auditQueueSize))
	}
	if c.AuditURL != "" {
		observers = append(observers, audit.NewAsyncObserver(audit.NewHTTPObserver(c.AuditURL), auditQueueSize))
	}

	if len(observers) == 0 {
		return nil, nil
	}

	p := audit.NewPublisher(observers...)

	return p, p
}
