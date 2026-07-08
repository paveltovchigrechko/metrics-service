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

	err = http.ListenAndServe(cfg.ServerAddress, r)
	if err != nil {
		log.Fatal(err)
	}
}
