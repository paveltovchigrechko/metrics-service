package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/paveltovchigrechko/metrics-service/internal/agent"
	"github.com/paveltovchigrechko/metrics-service/internal/buildinfo"
	"github.com/paveltovchigrechko/metrics-service/internal/config"
)

// Build information, set at build time:
//
//	go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=$(date +%F) -X main.buildCommit=$(git rev-parse --short HEAD)"
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	buildinfo.Print(os.Stdout, buildVersion, buildDate, buildCommit)

	cfg, err := config.SetAgentConfig(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	agent := agent.NewAgent(cfg)
	agent.Run(ctx)
}
