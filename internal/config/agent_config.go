package config

import (
	"errors"
	"flag"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	ServerAddress  string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

type envAgentConfig struct {
	ServerAddress  *string `env:"ADDRESS"`
	ReportInterval *int    `env:"REPORT_INTERVAL"`
	PollInterval   *int    `env:"POLL_INTERVAL"`
}

const (
	servAddressFlag       = "a"
	reportIntervalFlag    = "r"
	pollIntervalFlag      = "p"
	defaultServerAddress  = "localhost:8080"
	defaultReportInterval = 10
	defaultPollInterval   = 2
)

var errIncorrectInterval = errors.New("interval must be positive") // Make this error more descriptive: add flag and value that caused it.

func SetAgentConfig(args []string) (*AgentConfig, error) {
	envCfg, err := createAgentEnvConfig()
	if err != nil {
		return nil, err
	}

	flagCfg, err := createAgentFlagConfig(args)
	if err != nil {
		return nil, err
	}

	cfg := mergeAgentConfigs(envCfg, flagCfg)

	if err := validateIntervals(cfg.PollInterval, cfg.ReportInterval); err != nil {
		return nil, err
	}

	return cfg, nil
}

func mergeAgentConfigs(envCfg *envAgentConfig, flagCfg *AgentConfig) *AgentConfig {
	// Set server address
	if envCfg.ServerAddress != nil {
		flagCfg.ServerAddress = *envCfg.ServerAddress
	}

	// Set report interval
	if envCfg.ReportInterval != nil {
		reportIntervalSeconds := time.Duration(*envCfg.ReportInterval) * time.Second
		flagCfg.ReportInterval = reportIntervalSeconds
	}

	// Set poll interval
	if envCfg.PollInterval != nil {
		pollIntervalSeconds := time.Duration(*envCfg.PollInterval) * time.Second
		flagCfg.PollInterval = pollIntervalSeconds
	}

	return flagCfg
}

// createAgentFlagConfig parses the arguments and returns AgentConfig. The function doesn't validate parsed values.
func createAgentFlagConfig(args []string) (*AgentConfig, error) {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	address := fs.String(servAddressFlag, defaultServerAddress, "Metrics server HTTP address")
	reportSeconds := fs.Int(reportIntervalFlag, defaultReportInterval, "Metrics report frequency (seconds)")
	pollSeconds := fs.Int(pollIntervalFlag, defaultPollInterval, "Metrics update frequency (seconds)")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	reportInterval := time.Duration(*reportSeconds) * time.Second
	pollInterval := time.Duration(*pollSeconds) * time.Second

	return &AgentConfig{
		ServerAddress:  *address,
		ReportInterval: reportInterval,
		PollInterval:   pollInterval,
	}, nil
}

// createAgentEnvCongig parses the environment variables, validates string values for empty values, and returns envAgentConfig.
func createAgentEnvConfig() (*envAgentConfig, error) {
	envCfg := &envAgentConfig{}

	err := env.Parse(envCfg)
	if err != nil {
		return nil, err
	}

	// Treat empty environment variables as error
	if envCfg.ServerAddress != nil && strings.Trim(*envCfg.ServerAddress, " ") == "" {
		return nil, errEmptyServerAddress
	}

	return envCfg, nil
}

func validateIntervals(rInt, pInt time.Duration) error {
	if rInt <= 0 || pInt <= 0 {
		return errIncorrectInterval
	}

	return nil
}
