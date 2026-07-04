package config

import (
	"errors"
	"flag"
	"time"
)

type AgentConfig struct {
	ServerAddress  string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

const (
	servAddressFlag       = "a"
	reportIntervalFlag    = "r"
	pollIntervalFlag      = "p"
	defaultServerAddress  = "localhost:8080"
	defaultReportInterval = 10
	defaultPollInterval   = 2
)

var ErrIncorrectInterval = errors.New("interval must be positive")

func newAgentConfig(addr string, repInt, pollInt time.Duration) *AgentConfig {
	return &AgentConfig{
		ServerAddress:  addr,
		ReportInterval: repInt,
		PollInterval:   pollInt,
	}
}

func SetAgentConfig(args []string) (*AgentConfig, error) {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	address := fs.String(servAddressFlag, defaultServerAddress, "Metrics server HTTP address")
	reportSeconds := fs.Int(reportIntervalFlag, defaultReportInterval, "Metrics report frequency (seconds)")
	pollSeconds := fs.Int(pollIntervalFlag, defaultPollInterval, "Metrics update frequency (seconds)")

	err := fs.Parse(args)

	if err != nil {
		return nil, err
	}
	if err := validateIntervals(*reportSeconds, *pollSeconds); err != nil {
		return nil, err
	}

	reportInterval := time.Duration(*reportSeconds) * time.Second
	pollInterval := time.Duration(*pollSeconds) * time.Second
	return newAgentConfig(*address, reportInterval, pollInterval), nil
}

func validateIntervals(rInt, pInt int) error {
	if rInt <= 0 || pInt <= 0 {
		return ErrIncorrectInterval
	}

	return nil
}
