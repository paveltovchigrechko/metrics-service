package main

import (
	"log"

	"github.com/paveltovchigrechko/metrics-service/internal/agent"
)

func main() {
	agent, err := agent.NewAgent()
	if err != nil {
		log.Fatal(err)
	}
	agent.Run()
}
