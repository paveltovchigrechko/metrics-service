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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()
	agent, err := agent.NewAgent(cfg)
	if err != nil {
		log.Fatal(err)
	}
	agent.Run(ctx)
}
