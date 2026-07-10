package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/paveltovchigrechko/metrics-service/internal/config"
	"github.com/paveltovchigrechko/metrics-service/internal/logger"
	"github.com/paveltovchigrechko/metrics-service/internal/router"
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
	defer l.Sync()

	r := chi.NewRouter()
	r.Use(logger.Middleware(l))
	router.SetRoutes(r)

	err = http.ListenAndServe(cfg.ServerAddress, r)
	if err != nil {
		log.Fatal(err)
	}
}
