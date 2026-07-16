package config

import (
	"errors"
	"flag"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
)

// Default constants for command line flags.
const (
	serverAddressFlag   = "a"
	storeIntervalFlag   = "i"
	fileStoragePathFlag = "f"
	restoreFlag         = "r"

	defaultAddress         = "localhost:8080"
	defaultStoreInterval   = 300
	defaultFileStoragePath = "metrics.json"
)

var (
	errEmptyFileStoragePath = errors.New("environment variable for file storage path is an empty string")
	errEmptyServerAddress   = errors.New("environment variable for server address is an empty string")
)

// ServerConfig holds the server parameters.
type ServerConfig struct {
	ServerAddress   string        // Address with port to serve the server.
	StoreInterval   time.Duration // Interval for storing the server metrics. If value is 0, server saves metrics after each update.
	FileStoragePath string        // Local file address to store server metrics.
	Restore         bool          // Indicates if the server restores the metrics saved on FileStoragePath on start.
}

// envServerCongig stores the environment variables. All fields are pointers to distinguish if a value was set.
type envServerConfig struct {
	ServerAddress   *string `env:"ADDRESS"`
	StoreInterval   *int    `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
}

// SetSeverConfig returns the final ServerConfig to use.
// The function uses environment and flag configurations to merge into the final one.
func SetServerConfig(args []string) (*ServerConfig, error) {
	envCfg, err := createEnvConfig()
	if err != nil {
		return nil, err
	}

	flagCfg, err := createServerFlagConfig(args)
	if err != nil {
		return nil, err
	}

	cfg := mergeConfigs(envCfg, flagCfg)

	// Validate interval, 0 is valid here.
	if cfg.StoreInterval < 0 {
		return nil, errIncorrectInterval
	}

	return cfg, nil
}

// mergeConfigs merges envServerConfig and ServerConfig with environment configuration precedence.
func mergeConfigs(envCfg *envServerConfig, flagCfg *ServerConfig) *ServerConfig {
	// Set server address
	if envCfg.ServerAddress != nil {
		flagCfg.ServerAddress = *envCfg.ServerAddress
	}

	// Set store interval
	if envCfg.StoreInterval != nil {
		storeIntervalSeconds := time.Duration(*envCfg.StoreInterval) * time.Second
		flagCfg.StoreInterval = storeIntervalSeconds
	}

	// Set file path to store metrics
	if envCfg.FileStoragePath != nil {
		flagCfg.FileStoragePath = *envCfg.FileStoragePath
	}

	// Set restore option
	if envCfg.Restore != nil {
		flagCfg.Restore = *envCfg.Restore
	}

	return flagCfg
}

// createServerFlagConfig parses the arguments and returns ServerConfig. The function doesn't validate parsed values.
func createServerFlagConfig(args []string) (*ServerConfig, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	address := fs.String(serverAddressFlag, defaultAddress, "Metrics server HTTP address and port")
	storeInterval := fs.Int(storeIntervalFlag, defaultStoreInterval, "Metrics store interval in seconds")
	fileStoragePath := fs.String(fileStoragePathFlag, defaultFileStoragePath, "File name to store metrics")
	restore := fs.Bool(restoreFlag, true, "Indicate if server should restore metrics from the storage file")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	storeIntervalSec := time.Duration(*storeInterval) * time.Second

	cfg := &ServerConfig{
		ServerAddress:   *address,
		StoreInterval:   storeIntervalSec,
		FileStoragePath: *fileStoragePath,
		Restore:         *restore,
	}

	return cfg, nil
}

// createEnvCongig parses the environment variables, validates string values for empty values, and returns envServerConfig.
func createEnvConfig() (*envServerConfig, error) {
	envCfg := &envServerConfig{}

	err := env.Parse(envCfg)
	if err != nil {
		return nil, err
	}

	// Treat empty environment variables as error
	if envCfg.FileStoragePath != nil && strings.Trim(*envCfg.FileStoragePath, " ") == "" {
		return nil, errEmptyFileStoragePath
	}

	if envCfg.ServerAddress != nil && strings.Trim(*envCfg.ServerAddress, " ") == "" {
		return nil, errEmptyServerAddress
	}

	return envCfg, nil
}
