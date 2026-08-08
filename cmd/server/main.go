package main

import (
	"log"
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

	serv, err := server.New(cfg,
		logger.LoggerMiddleware(l),
		middleware.VerifyHashMiddleware(cfg.Key),
		middleware.SignResponseMiddleware(cfg.Key),
		middleware.GZIPMiddleware,
	)
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
