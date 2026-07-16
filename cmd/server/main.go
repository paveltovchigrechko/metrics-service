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
	run()
}

func run() {
	cfg, err := config.SetServerConfig(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	l, err := logger.New()
	if err != nil {
		log.Fatal(err)
	}
	defer l.Sync() // Flush at the end

	serv := server.NewServer(cfg)
	serv.UseMiddlewares(logger.LoggerMiddleware(l), middleware.GZIPMiddleware) // Set middlewares before setting handlers

	err = serv.Run()
	if err != nil {
		log.Fatal(err)
	}
}
