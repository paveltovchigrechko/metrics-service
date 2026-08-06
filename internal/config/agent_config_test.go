package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAgentConfig_Success(t *testing.T) {
	testCases := []struct {
		name       string
		args       []string
		envVars    map[string]string
		wantAddr   string
		wantReport time.Duration
		wantPoll   time.Duration
		wantKey    string
	}{
		{
			name:       "Default values when no env or flags are set",
			args:       []string{},
			envVars:    map[string]string{},
			wantAddr:   "localhost:8080",
			wantReport: 10 * time.Second,
			wantPoll:   2 * time.Second,
			wantKey:    "",
		},
		{
			name:       "Flags are parsed correctly with no env present",
			args:       []string{"-a", "127.0.0.1:9090", "-r", "30", "-p", "5", "-k", "key"},
			envVars:    map[string]string{},
			wantAddr:   "127.0.0.1:9090",
			wantReport: 30 * time.Second,
			wantPoll:   5 * time.Second,
			wantKey:    "key",
		},
		{
			name: "Env variables completely override flag configurations",
			args: []string{"-a", "127.0.0.1:9090", "-r", "30", "-p", "5", "-k", "key"},
			envVars: map[string]string{
				"ADDRESS":         "0.0.0.0:3000",
				"REPORT_INTERVAL": "40",
				"POLL_INTERVAL":   "10",
				"KEY":             "env key",
			},
			wantAddr:   "0.0.0.0:3000",
			wantReport: 40 * time.Second,
			wantPoll:   10 * time.Second,
			wantKey:    "env key",
		},
		{
			name: "Partial Env overrides only address",
			args: []string{"-a", "127.0.0.1:9090", "-r", "30", "-p", "5", "-k", "key"},
			envVars: map[string]string{
				"ADDRESS": "0.0.0.0:3000",
			},
			wantAddr:   "0.0.0.0:3000",
			wantReport: 30 * time.Second,
			wantPoll:   5 * time.Second,
			wantKey:    "key",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear testing sandbox envs
			t.Setenv("ADDRESS", "")
			t.Setenv("REPORT_INTERVAL", "")
			t.Setenv("POLL_INTERVAL", "")
			t.Setenv("KEY", "")

			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}

			cfg, err := SetAgentConfig(tc.args)
			require.NoError(t, err)
			require.NotNil(t, cfg)

			assert.Equal(t, tc.wantAddr, cfg.ServerAddress)
			assert.Equal(t, tc.wantReport, cfg.ReportInterval)
			assert.Equal(t, tc.wantPoll, cfg.PollInterval)
		})
	}
}

func TestSetAgentConfig_Failures(t *testing.T) {
	t.Run("fails on invalid flag structure", func(t *testing.T) {
		cfg, err := SetAgentConfig([]string{"-invalid-arg"})
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("fails on invalid environment address with validation error", func(t *testing.T) {
		t.Setenv("ADDRESS", "   ") // empty address check
		cfg, err := SetAgentConfig([]string{})
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("fails when environment variable types are mismatched", func(t *testing.T) {
		t.Setenv("REPORT_INTERVAL", "invalid-type")
		cfg, err := SetAgentConfig([]string{})
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("fails on negative/zero report interval with descriptive error", func(t *testing.T) {
		cfg, err := SetAgentConfig([]string{"-r", "0"})
		require.Error(t, err)
		assert.ErrorIs(t, err, errIncorrectInterval)
		assert.Nil(t, cfg)
	})

	t.Run("fails on negative/zero poll interval with descriptive error", func(t *testing.T) {
		cfg, err := SetAgentConfig([]string{"-p", "-3"})
		require.Error(t, err)
		assert.ErrorIs(t, err, errIncorrectInterval)
		assert.Nil(t, cfg)
	})
}

func TestCreateAgentFlagConfig(t *testing.T) {
	t.Run("should correctly parse flag options into agent config struct", func(t *testing.T) {
		args := []string{"-a", "10.0.0.1:1337", "-r", "60", "-p", "12", "-k", "key"}
		cfg, err := createAgentFlagConfig(args)

		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, "10.0.0.1:1337", cfg.ServerAddress)
		assert.Equal(t, 60*time.Second, cfg.ReportInterval)
		assert.Equal(t, 12*time.Second, cfg.PollInterval)
		assert.Equal(t, "key", cfg.Key)
	})
}

func TestCreateAgentEnvConfig(t *testing.T) {
	t.Run("parses non-empty env configs cleanly", func(t *testing.T) {
		t.Setenv("ADDRESS", "192.168.1.1:80")
		t.Setenv("REPORT_INTERVAL", "15")
		t.Setenv("POLL_INTERVAL", "3")
		t.Setenv("KEY", "env key")

		envCfg, err := createAgentEnvConfig()
		require.NoError(t, err)
		require.NotNil(t, envCfg)

		assert.Equal(t, "192.168.1.1:80", *envCfg.ServerAddress)
		assert.Equal(t, 15, *envCfg.ReportInterval)
		assert.Equal(t, 3, *envCfg.PollInterval)
		assert.Equal(t, "env key", *envCfg.Key)
	})

	t.Run("avoids pointer dereferences on empty environment (returns nils)", func(t *testing.T) {
		t.Setenv("ADDRESS", "")
		t.Setenv("REPORT_INTERVAL", "")
		t.Setenv("POLL_INTERVAL", "")
		t.Setenv("KEY", "")

		envCfg, err := createAgentEnvConfig()
		require.NoError(t, err)
		require.NotNil(t, envCfg)

		assert.Nil(t, envCfg.ServerAddress)
		assert.Nil(t, envCfg.ReportInterval)
		assert.Nil(t, envCfg.PollInterval)
		assert.Nil(t, envCfg.Key)
	})
}

func TestMergeAgentConfigs(t *testing.T) {
	strPtr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }

	t.Run("overwrites config with env parameters when pointers are filled", func(t *testing.T) {
		flagCfg := &AgentConfig{
			ServerAddress:  "flag:80",
			ReportInterval: 5 * time.Second,
			PollInterval:   1 * time.Second,
			Key:            "key",
		}

		envCfg := &envAgentConfig{
			ServerAddress:  strPtr("env:80"),
			ReportInterval: intPtr(100),
			PollInterval:   intPtr(20),
			Key:            strPtr("env key"),
		}

		merged := mergeAgentConfigs(envCfg, flagCfg)
		require.NotNil(t, merged)
		assert.Equal(t, "env:80", merged.ServerAddress)
		assert.Equal(t, 100*time.Second, merged.ReportInterval)
		assert.Equal(t, 20*time.Second, merged.PollInterval)
		assert.Equal(t, "env key", merged.Key)
	})
}
