package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetServerConfig_Positive(t *testing.T) {
	testCases := []struct {
		name         string
		args         []string
		envVars      map[string]string
		wantAddr     string
		wantInterval time.Duration
		wantPath     string
		wantRestore  bool
	}{
		{
			name:         "Default values when no env or flags are set",
			args:         []string{},
			envVars:      map[string]string{},
			wantAddr:     "localhost:8080",
			wantInterval: 300 * time.Second,
			wantPath:     "metrics.json",
			wantRestore:  true,
		},
		{
			name:         "Flags are parsed correctly when no env is present",
			args:         []string{"-a", "127.0.0.1:9090", "-i", "10", "-f", "test.json", "-r=false"},
			envVars:      map[string]string{},
			wantAddr:     "127.0.0.1:9090",
			wantInterval: 10 * time.Second,
			wantPath:     "test.json",
			wantRestore:  false,
		},
		{
			name: "Env variables override flags completely",
			args: []string{"-a", "127.0.0.1:9090", "-i", "10", "-f", "test.json", "-r=false"},
			envVars: map[string]string{
				"ADDRESS":           "0.0.0.0:3000",
				"STORE_INTERVAL":    "15",
				"FILE_STORAGE_PATH": "env.json",
				"RESTORE":           "true",
			},
			wantAddr:     "0.0.0.0:3000",
			wantInterval: 15 * time.Second,
			wantPath:     "env.json",
			wantRestore:  true,
		},
		{
			name: "Partial Env overrides only specific flags",
			args: []string{"-a", "127.0.0.1:9090", "-i", "10"},
			envVars: map[string]string{
				"ADDRESS": "0.0.0.0:3000",
			},
			wantAddr:     "0.0.0.0:3000",
			wantInterval: 10 * time.Second,
			wantPath:     "metrics.json",
			wantRestore:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean environments first, then set test case specific envs
			t.Setenv("ADDRESS", "")
			t.Setenv("STORE_INTERVAL", "")
			t.Setenv("FILE_STORAGE_PATH", "")
			t.Setenv("RESTORE", "")

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
