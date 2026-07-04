package main

import (
	"log"
	"os"

	"github.com/paveltovchigrechko/metrics-service/internal/agent"
	"github.com/paveltovchigrechko/metrics-service/internal/config"
)

func main() {
	cfg, err := config.SetAgentConfig(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	agent := agent.NewAgent(cfg)
	agent.Run()
}
