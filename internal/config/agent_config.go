package config

import (
	"errors"
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	ServerAddress  string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

type envConfig struct {
	ServerAddress  string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

const (
	servAddressFlag       = "a"
	reportIntervalFlag    = "r"
	pollIntervalFlag      = "p"
	defaultServerAddress  = "localhost:8080"
	defaultReportInterval = 10
	defaultPollInterval   = 2
)

var errIncorrectInterval = errors.New("interval must be positive")

func newAgentConfig(addr string, repInt, pollInt time.Duration) *AgentConfig {
	return &AgentConfig{
		ServerAddress:  addr,
		ReportInterval: repInt,
		PollInterval:   pollInt,
	}
}

func SetAgentConfig(args []string) (*AgentConfig, error) {
	// To think about: the priority, currently the errors in flag values breaks the configuration biuld even if env variables set correctly.
	flagConfig, err := createFlagConfig(args)
	if err != nil {
		return nil, err
	}

	cfg, err := createConfig(flagConfig)
	if err != nil {
		return nil, err
	}

	log.Printf("Set server address to %s\n", cfg.ServerAddress)
	log.Printf("Set report interval to %v\n", cfg.ReportInterval)
	log.Printf("Set poll interval to %v\n", cfg.PollInterval)
	return cfg, nil
}

func createConfig(cfg *AgentConfig) (*AgentConfig, error) {
	envConfig := envConfig{}
	err := readEnvVariables(&envConfig)
	if err != nil {
		return nil, err
	}

	if envConfig.ServerAddress != "" {
		cfg.ServerAddress = envConfig.ServerAddress
	}
	if envConfig.ReportInterval != 0 {
		cfg.ReportInterval = time.Duration(envConfig.ReportInterval) * time.Second
	}
	if envConfig.PollInterval != 0 {
		cfg.PollInterval = time.Duration(envConfig.PollInterval) * time.Second
	}

	return cfg, nil
}

func createFlagConfig(args []string) (*AgentConfig, error) {
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

func readEnvVariables(cfg *envConfig) error {
	err := env.Parse(cfg)
	if err != nil {
		return err
	}

	return nil
}

func validateIntervals(rInt, pInt int) error {
	if rInt <= 0 || pInt <= 0 {
		return errIncorrectInterval
	}

	return nil
}
