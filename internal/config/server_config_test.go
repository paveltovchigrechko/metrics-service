package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetServerConfig_Positive(t *testing.T) {
	testCases := []struct {
		name              string
		args              []string
		envVars           map[string]string
		wantAddr          string
		wantInterval      time.Duration
		wantPath          string
		wantRestore       bool
		wantDSN           string
		wantKey           string
		wantAuditFilePath string
		wantAuditURL      string
	}{
		{
			name:              "Default values when no env or flags are set",
			args:              []string{},
			envVars:           map[string]string{},
			wantAddr:          "localhost:8080",
			wantInterval:      300 * time.Second,
			wantPath:          "metrics.json",
			wantRestore:       true,
			wantDSN:           "",
			wantKey:           "",
			wantAuditFilePath: "",
			wantAuditURL:      "",
		},
		{
			name:              "Flags are parsed correctly when no env is present",
			args:              []string{"-a", "127.0.0.1:9090", "-i", "10", "-f", "test.json", "-r=false", "-d", "db_address", "-k", "key", "--audit-file", "audit-file-path", "--audit-url", "audit-url"},
			envVars:           map[string]string{},
			wantAddr:          "127.0.0.1:9090",
			wantInterval:      10 * time.Second,
			wantPath:          "test.json",
			wantRestore:       false,
			wantDSN:           "db_address",
			wantKey:           "key",
			wantAuditFilePath: "audit-file-path",
			wantAuditURL:      "audit-url",
		},
		{
			name: "Env variables override flags completely",
			args: []string{"-a", "127.0.0.1:9090", "-i", "10", "-f", "test.json", "-r=false", "-d", "db_address", "-k", "key", "--audit-file", "audit-file-path", "--audit-url", "audit-url"},
			envVars: map[string]string{
				"ADDRESS":           "0.0.0.0:3000",
				"STORE_INTERVAL":    "15",
				"FILE_STORAGE_PATH": "env.json",
				"RESTORE":           "true",
				"DATABASE_DSN":      "env_db_address",
				"KEY":               "env key",
				"AUDIT_FILE":        "env_audit-file-path",
				"AUDIT_URL":         "env_audit-url",
			},
			wantAddr:          "0.0.0.0:3000",
			wantInterval:      15 * time.Second,
			wantPath:          "env.json",
			wantRestore:       true,
			wantDSN:           "env_db_address",
			wantKey:           "env key",
			wantAuditFilePath: "env_audit-file-path",
			wantAuditURL:      "env_audit-url",
		},
		{
			name: "Partial Env overrides only specific flags",
			args: []string{"-a", "127.0.0.1:9090", "-i", "10", "-d", "flag_db_address", "-k", "key", "--audit-file", "audit-file-path", "--audit-url", "audit-url"},
			envVars: map[string]string{
				"ADDRESS": "0.0.0.0:3000",
			},
			wantAddr:          "0.0.0.0:3000",
			wantInterval:      10 * time.Second,
			wantPath:          "metrics.json",
			wantRestore:       true,
			wantDSN:           "flag_db_address",
			wantKey:           "key",
			wantAuditFilePath: "audit-file-path",
			wantAuditURL:      "audit-url",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean environments first, then set test case specific envs
			t.Setenv("ADDRESS", "")
			t.Setenv("STORE_INTERVAL", "")
			t.Setenv("FILE_STORAGE_PATH", "")
			t.Setenv("RESTORE", "")
			t.Setenv("DATABASE_DSN", "")
			t.Setenv("KEY", "")
			t.Setenv("AUDIT_FILE", "")
			t.Setenv("AUDIT_URL", "")

			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}

			cfg, err := SetServerConfig(tc.args)
			require.NoError(t, err)
			require.NotNil(t, cfg)

			assert.Equal(t, tc.wantAddr, cfg.ServerAddress)
			assert.Equal(t, tc.wantInterval, cfg.StoreInterval)
			assert.Equal(t, tc.wantPath, cfg.FileStoragePath)
			assert.Equal(t, tc.wantRestore, cfg.Restore)
			assert.Equal(t, tc.wantDSN, cfg.DatabaseDSN)
			assert.Equal(t, tc.wantKey, cfg.Key)
			assert.Equal(t, tc.wantAuditFilePath, cfg.AuditFilePath)
			assert.Equal(t, tc.wantAuditURL, cfg.AuditURL)
		})
	}
}

func TestSetServerConfig_Negative(t *testing.T) {
	t.Run("returns error when an invalid flag is passed", func(t *testing.T) {
		cfg, err := SetServerConfig([]string{"-invalid-flag-structure"})
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("returns error when interval is negative", func(t *testing.T) {
		cfg, err := SetServerConfig([]string{"-i", "-5"})
		assert.ErrorIs(t, err, errIncorrectInterval)
		assert.Nil(t, cfg)
	})

	t.Run("returns error when empty environment ADDRESS is passed", func(t *testing.T) {
		t.Setenv("ADDRESS", " ")
		cfg, err := SetServerConfig([]string{})
		assert.ErrorIs(t, err, errEmptyServerAddress)
		assert.Nil(t, cfg)
	})

	t.Run("returns error when empty environment FILE_STORAGE_PATH is passed", func(t *testing.T) {
		t.Setenv("ADDRESS", "localhost:8080")
		t.Setenv("FILE_STORAGE_PATH", "   ")
		cfg, err := SetServerConfig([]string{})
		assert.ErrorIs(t, err, errEmptyFileStoragePath)
		assert.Nil(t, cfg)
	})

	t.Run("returns error when STORE_INTERVAL env contains invalid type", func(t *testing.T) {
		t.Setenv("STORE_INTERVAL", "not-an-integer")
		cfg, err := SetServerConfig([]string{})
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}

func TestCreateServerFlagConfig(t *testing.T) {
	t.Run("should parse custom flags successfully", func(t *testing.T) {
		args := []string{"-a", "127.0.0.1:4444", "-i", "120", "-f", "flag_test.json", "-r=false", "-d", "db_address", "-k", "key", "--audit-file", "audit-file-path", "--audit-url", "audit-url"}
		cfg, err := createServerFlagConfig(args)

		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, "127.0.0.1:4444", cfg.ServerAddress)
		assert.Equal(t, 120*time.Second, cfg.StoreInterval)
		assert.Equal(t, "flag_test.json", cfg.FileStoragePath)
		assert.False(t, cfg.Restore)
		assert.Equal(t, "db_address", cfg.DatabaseDSN)
		assert.Equal(t, "key", cfg.Key)
		assert.Equal(t, "audit-file-path", cfg.AuditFilePath)
		assert.Equal(t, "audit-url", cfg.AuditURL)
	})

	t.Run("should return error on invalid flags", func(t *testing.T) {
		args := []string{"-unsupported-flag"}
		cfg, err := createServerFlagConfig(args)

		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}

func TestCreateEnvConfig(t *testing.T) {
	t.Run("should parse valid environment variables", func(t *testing.T) {
		t.Setenv("ADDRESS", "127.0.0.1:3030")
		t.Setenv("STORE_INTERVAL", "45")
		t.Setenv("FILE_STORAGE_PATH", "env_test.json")
		t.Setenv("RESTORE", "false")
		t.Setenv("DATABASE_DSN", "env_db_address")
		t.Setenv("KEY", "env_key")
		t.Setenv("AUDIT_FILE", "env_audit-file")
		t.Setenv("AUDIT_URL", "env_audit-url")

		envCfg, err := parseEnvServerConfig()

		require.NoError(t, err)
		require.NotNil(t, envCfg)

		assert.Equal(t, "127.0.0.1:3030", *envCfg.ServerAddress)
		assert.Equal(t, 45, *envCfg.StoreInterval)
		assert.Equal(t, "env_test.json", *envCfg.FileStoragePath)
		assert.False(t, *envCfg.Restore)
		assert.Equal(t, "env_db_address", *envCfg.DatabaseDSN)
		assert.Equal(t, "env_key", *envCfg.Key)
		assert.Equal(t, "env_audit-file", *envCfg.AuditFilePath)
		assert.Equal(t, "env_audit-url", *envCfg.AuditURL)
	})

	t.Run("should succeed when no env vars are defined (pointers are nil)", func(t *testing.T) {
		t.Setenv("ADDRESS", "")
		t.Setenv("STORE_INTERVAL", "")
		t.Setenv("FILE_STORAGE_PATH", "")
		t.Setenv("RESTORE", "")
		t.Setenv("DATABASE_DSN", "")
		t.Setenv("KEY", "")
		t.Setenv("AUDIT_FILE", "")
		t.Setenv("AUDIT_URL", "")

		envCfg, err := parseEnvServerConfig()

		require.NoError(t, err)
		require.NotNil(t, envCfg)
		assert.Nil(t, envCfg.ServerAddress)
		assert.Nil(t, envCfg.StoreInterval)
		assert.Nil(t, envCfg.FileStoragePath)
		assert.Nil(t, envCfg.Restore)
		assert.Nil(t, envCfg.DatabaseDSN)
		assert.Nil(t, envCfg.Key)
		assert.Nil(t, envCfg.AuditFilePath)
		assert.Nil(t, envCfg.AuditURL)
	})
}

func TestMergeConfigs(t *testing.T) {
	t.Run("should overwrite flag config with env values", func(t *testing.T) {
		flagCfg := &ServerConfig{
			ServerAddress:   "flag-address:8080",
			StoreInterval:   10 * time.Second,
			FileStoragePath: "flag-file.json",
			Restore:         true,
			DatabaseDSN:     "flag_db_address",
			Key:             "key",
			AuditFilePath:   "audit-file-path",
			AuditURL:        "audit-url",
		}

		envCfg := &envServerConfig{
			ServerAddress:   new("env-address:9090"),
			StoreInterval:   new(50),
			FileStoragePath: new("env-file.json"),
			Restore:         new(false),
			DatabaseDSN:     new("env_db_address"),
			Key:             new("env_key"),
			AuditFilePath:   new("env_audit-file-path"),
			AuditURL:        new("env_audit-url"),
		}

		merged := mergeServerConfigs(envCfg, flagCfg)

		require.NotNil(t, merged)
		assert.Equal(t, "env-address:9090", merged.ServerAddress)
		assert.Equal(t, 50*time.Second, merged.StoreInterval)
		assert.Equal(t, "env-file.json", merged.FileStoragePath)
		assert.False(t, merged.Restore)
		assert.Equal(t, "env_db_address", merged.DatabaseDSN)
		assert.Equal(t, "env_key", merged.Key)
		assert.Equal(t, "env_audit-file-path", merged.AuditFilePath)
		assert.Equal(t, "env_audit-url", merged.AuditURL)
	})

	t.Run("should retain flag config values if env pointers are nil", func(t *testing.T) {
		flagCfg := &ServerConfig{
			ServerAddress:   "flag-address:8080",
			StoreInterval:   10 * time.Second,
			FileStoragePath: "flag-file.json",
			Restore:         true,
			DatabaseDSN:     "flag_db_address",
			Key:             "key",
			AuditFilePath:   "audit-file-path",
			AuditURL:        "audit-url",
		}

		envCfg := &envServerConfig{
			ServerAddress:   nil,
			StoreInterval:   nil,
			FileStoragePath: nil,
			Restore:         nil,
			DatabaseDSN:     nil,
			Key:             nil,
			AuditFilePath:   nil,
			AuditURL:        nil,
		}

		merged := mergeServerConfigs(envCfg, flagCfg)

		require.NotNil(t, merged)
		assert.Equal(t, "flag-address:8080", merged.ServerAddress)
		assert.Equal(t, 10*time.Second, merged.StoreInterval)
		assert.Equal(t, "flag-file.json", merged.FileStoragePath)
		assert.True(t, merged.Restore)
		assert.Equal(t, "flag_db_address", merged.DatabaseDSN)
		assert.Equal(t, "key", merged.Key)
		assert.Equal(t, "audit-file-path", merged.AuditFilePath)
		assert.Equal(t, "audit-url", merged.AuditURL)
	})
}
