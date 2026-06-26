package main

import (
	"github.com/paveltovchigrechko/metrics-service/internal/agent"
)

func main() {
	agent := agent.NewAgent()
	agent.Run()
}
