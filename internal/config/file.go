package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const (
	configFlagShort = "c"
	configFlagLong  = "config"
)

// fileAgentConfig is the agent JSON file. Values are Pointers to distinct from an explicit zero value.
type fileAgentConfig struct {
	Address        *string `json:"address"`
	ReportInterval *string `json:"report_interval"`
	PollInterval   *string `json:"poll_interval"`
	Key            *string `json:"key"`
	RateLimit      *int    `json:"rate_limit"`
	CryptoKey      *string `json:"crypto_key"`
}

// fileServerConfig is the server JSON file. Values are Pointers to distinct from an explicit zero value.
type fileServerConfig struct {
	Address       *string `json:"address"`        
	Restore       *bool   `json:"restore"`       
	StoreInterval *string `json:"store_interval"` 
	StoreFile     *string `json:"store_file"`     
	DatabaseDSN   *string `json:"database_dsn"`   
	Key           *string `json:"key"`            
	AuditFile     *string `json:"audit_file"`     
	AuditURL      *string `json:"audit_url"`      
	CryptoKey     *string `json:"crypto_key"`     
}

// configFilePath returns the JSON file path. CONFIG env variable overrides -c/-config.
// An empty string means the file is not used.
func configFilePath(flagPath string, envPath *string) string {
	if envPath != nil {
		return strings.TrimSpace(*envPath)
	}
	return strings.TrimSpace(flagPath)
}

// loadFileConfig decodes the JSON file at path into dest. An empty path
// leaves dest unchanged. Unknown fields and trailing values are rejected.
func loadFileConfig(path string, dest any) error {
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %s: %w", path, err)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dest); err != nil {
		return fmt.Errorf("parse config file %s: %w", path, err)
	}

	var trailing any
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		return fmt.Errorf("parse config file %s: extra data after JSON value", path)
	}

	return nil
}

func parseFileDuration(name, raw string) (time.Duration, error) {
	d, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("config %s: %w", name, err)
	}
	return d, nil
}

// applyAgentFileConfig copies file values onto cfg. A value is skipped when
// the matching flag was set on the command line, so flags beat the file.
// Environment variables are applied later and beat both.
func applyAgentFileConfig(cfg *AgentConfig, file *fileAgentConfig, explicit map[string]bool) error {
	if file.Address != nil && !explicit[servAddressFlag] {
		cfg.ServerAddress = strings.TrimSpace(*file.Address)
	}

	if file.ReportInterval != nil {
		d, err := parseFileDuration("report_interval", *file.ReportInterval)
		if err != nil {
			return err
		}
		if !explicit[reportIntervalFlag] {
			cfg.ReportInterval = d
		}
	}

	if file.PollInterval != nil {
		d, err := parseFileDuration("poll_interval", *file.PollInterval)
		if err != nil {
			return err
		}
		if !explicit[pollIntervalFlag] {
			cfg.PollInterval = d
		}
	}

	if file.Key != nil && !explicit[keyFlag] {
		cfg.Key = *file.Key
	}

	if file.RateLimit != nil && !explicit[rateLimitFlag] {
		cfg.RateLimit = *file.RateLimit
	}

	if file.CryptoKey != nil && !explicit[cryptoKeyFlag] {
		cfg.CryptoKey = strings.TrimSpace(*file.CryptoKey)
	}

	return nil
}

// applyServerFileConfig copies file values onto cfg.
func applyServerFileConfig(cfg *ServerConfig, file *fileServerConfig, explicit map[string]bool) error {
	if file.Address != nil && !explicit[serverAddressFlag] {
		cfg.ServerAddress = strings.TrimSpace(*file.Address)
	}

	if file.Restore != nil && !explicit[restoreFlag] {
		cfg.Restore = *file.Restore
	}

	if file.StoreInterval != nil {
		d, err := parseFileDuration("store_interval", *file.StoreInterval)
		if err != nil {
			return err
		}
		if !explicit[storeIntervalFlag] {
			cfg.StoreInterval = d
		}
	}

	if file.StoreFile != nil && !explicit[fileStoragePathFlag] {
		cfg.FileStoragePath = strings.TrimSpace(*file.StoreFile)
	}

	if file.DatabaseDSN != nil && !explicit[databaseDSNFlag] {
		cfg.DatabaseDSN = strings.TrimSpace(*file.DatabaseDSN)
	}

	if file.Key != nil && !explicit[secretKeyFlag] {
		cfg.Key = *file.Key
	}

	if file.AuditFile != nil && !explicit[auditFilePathFlag] {
		cfg.AuditFilePath = strings.TrimSpace(*file.AuditFile)
	}

	if file.AuditURL != nil && !explicit[auditURLFlag] {
		cfg.AuditURL = strings.TrimSpace(*file.AuditURL)
	}

	if file.CryptoKey != nil && !explicit[cryptoKeyFlag] {
		cfg.CryptoKey = strings.TrimSpace(*file.CryptoKey)
	}

	return nil
}
