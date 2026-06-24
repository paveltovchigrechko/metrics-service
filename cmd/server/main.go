package main

import (
	"log"
	"net/http"

	"github.com/paveltovchigrechko/metrics-service/internal/handler"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
)

func main() {
	run()
}

func run() {
	storage := models.NewStorage()
	server := http.NewServeMux()
	h := handler.NewHandler(storage)

	server.HandleFunc(`/`, h.MainPage)

	err := http.ListenAndServe(`:8080`, server)
	if err != nil {
		log.Fatal(err)
	}
}
