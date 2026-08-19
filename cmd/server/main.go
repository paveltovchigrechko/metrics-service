package main

import (
	"log"
	"net/http"
	"os"

	"github.com/paveltovchigrechko/metrics-service/internal/config"
	"github.com/paveltovchigrechko/metrics-service/internal/logger"
	"github.com/paveltovchigrechko/metrics-service/internal/middleware"
	"github.com/paveltovchigrechko/metrics-service/internal/server"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.SetServerConfig(os.Args[1:])
	if err != nil {
		return err
	}

	l, err := logger.New()
	if err != nil {
		return err
	}
	defer l.Sync() // Flush at the end

	middlewares := []func(http.Handler) http.Handler{
		logger.LoggerMiddleware(l),
	}

	if cfg.Key != "" {
		middlewares = append(middlewares,
			middleware.VerifyHashMiddleware(cfg.Key),
			middleware.SignResponseMiddleware(cfg.Key),
		)
	}

	middlewares = append(middlewares, middleware.GZIPMiddleware)

	serv, err := server.New(cfg, middlewares...)

	if err != nil {
		return err
	}
	defer serv.StopDB()

	err = serv.Run()
	if err != nil {
		return err
	}

	return nil
}
