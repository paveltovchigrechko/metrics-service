package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/paveltovchigrechko/metrics-service/internal/agent"
	"github.com/paveltovchigrechko/metrics-service/internal/config"
)

func main() {
	cfg, err := config.SetAgentConfig(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	agent := agent.NewAgent(cfg)
	agent.Run(ctx)
}
