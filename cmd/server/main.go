package main

import (
	"log"
	"net/http"

	"github.com/paveltovchigrechko/metrics-service/internal/router"
)

func main() {
	run()
}

func run() {
	r, cfg, err := router.PrepareServerRouterAndConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Starting server on %s", cfg.ServerAddress)
	err = http.ListenAndServe(cfg.ServerAddress, r)
	log.Printf("ListenAndServe returned: %v", err)
	if err != nil {
		log.Fatal(err)
	}
}
