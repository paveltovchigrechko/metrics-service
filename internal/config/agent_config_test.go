package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAgentConfig(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		envVars        map[string]string
		expectError    error
		expectedAddr   string
		expectedReport time.Duration
		expectedPoll   time.Duration
	}{
		{
			name:           "Scenario 1: Defaults are applied when no flags or envs are present",
			args:           []string{},
			envVars:        map[string]string{},
			expectError:    nil,
			expectedAddr:   "localhost:8080",
			expectedReport: 10 * time.Second,
			expectedPoll:   2 * time.Second,
		},
		{
			name:           "Scenario 2: Valid flags override defaults (No Env)",
			args:           []string{"-a", "127.0.0.1:9090", "-r", "30", "-p", "5"},
			envVars:        map[string]string{},
			expectError:    nil,
			expectedAddr:   "127.0.0.1:9090",
			expectedReport: 30 * time.Second,
			expectedPoll:   5 * time.Second,
		},
		{
			name:           "Scenario 3: Valid env variables override defaults (No Flags)",
			args:           []string{},
			envVars:        map[string]string{"ADDRESS": "0.0.0.0:3000", "REPORT_INTERVAL": "15", "POLL_INTERVAL": "4"},
			expectError:    nil,
			expectedAddr:   "0.0.0.0:3000",
			expectedReport: 15 * time.Second,
			expectedPoll:   4 * time.Second,
		},
		{
			name: "Scenario 4: Env variables take absolute priority over active flags",
			args: []string{"-a", "flag-host:1111", "-r", "50", "-p", "10"},
			envVars: map[string]string{
				"ADDRESS":         "env-host:2222",
				"REPORT_INTERVAL": "5",
				"POLL_INTERVAL":   "1",
			},
			expectError:    nil,
			expectedAddr:   "env-host:2222",
			expectedReport: 5 * time.Second,
			expectedPoll:   1 * time.Second,
		},
		{
			name:        "Scenario 5: Negative value on flag fails validation",
			args:        []string{"-r", "-10"},
			envVars:     map[string]string{},
			expectError: errIncorrectInterval,
		},
		{
			name:        "Scenario 6: Zero value on flag fails validation",
			args:        []string{"-p", "0"},
			envVars:     map[string]string{},
			expectError: errIncorrectInterval,
		},
		{
			name:        "Scenario 7: Malformed or unregistered flag returns parsing error",
			args:        []string{"-unknown-flag"},
			envVars:     map[string]string{},
			expectError: assert.AnError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for key, val := range tc.envVars {
				t.Setenv(key, val)
			}

			cfg, err := SetAgentConfig(tc.args)

			if tc.expectError != nil {
				assert.Error(t, err)
				if tc.expectError != assert.AnError {
					assert.ErrorIs(t, err, tc.expectError)
				}
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				assert.Equal(t, tc.expectedAddr, cfg.ServerAddress)
				assert.Equal(t, tc.expectedReport, cfg.ReportInterval)
				assert.Equal(t, tc.expectedPoll, cfg.PollInterval)
			}
		})
	}
}

func TestValidateIntervals(t *testing.T) {
	testCases := []struct {
		name      string
		reportInt int
		pollInt   int
		expectErr error
	}{
		{"valid positive intervals", 10, 2, nil},
		{"negative report interval", -5, 2, errIncorrectInterval},
		{"zero report interval", 0, 2, errIncorrectInterval},
		{"negative poll interval", 10, -2, errIncorrectInterval},
		{"zero poll interval", 10, 0, errIncorrectInterval},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateIntervals(tc.reportInt, tc.pollInt)
			if tc.expectErr != nil {
				assert.ErrorIs(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateFlagConfig(t *testing.T) {
	t.Run("successfully parses arguments into flag config", func(t *testing.T) {
		args := []string{"-a", "127.0.0.1:8081", "-r", "20", "-p", "4"}
		cfg, err := createFlagConfig(args)

		assert.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, "127.0.0.1:8081", cfg.ServerAddress)
		assert.Equal(t, 20*time.Second, cfg.ReportInterval)
		assert.Equal(t, 4*time.Second, cfg.PollInterval)
	})

	t.Run("returns error on parsing malformed flag options", func(t *testing.T) {
		args := []string{"-unknown-flag"}
		cfg, err := createFlagConfig(args)

		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("returns error on invalid parameter intervals", func(t *testing.T) {
		args := []string{"-r", "-10"}
		cfg, err := createFlagConfig(args)

		assert.ErrorIs(t, err, errIncorrectInterval)
		assert.Nil(t, cfg)
	})
}

func TestCreateConfig(t *testing.T) {
	t.Run("overwrites baseline configuration with env configurations", func(t *testing.T) {
		t.Setenv("ADDRESS", "localhost:9999")
		t.Setenv("REPORT_INTERVAL", "50")
		t.Setenv("POLL_INTERVAL", "5")

		baseCfg := &AgentConfig{
			ServerAddress:  "baseline:8080",
			ReportInterval: 10 * time.Second,
			PollInterval:   2 * time.Second,
		}

		cfg, err := createConfig(baseCfg)

		assert.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, "localhost:9999", cfg.ServerAddress)
		assert.Equal(t, 50*time.Second, cfg.ReportInterval)
		assert.Equal(t, 5*time.Second, cfg.PollInterval)
	})

	t.Run("retains original baseline settings if environment vars are omitted", func(t *testing.T) {
		baseCfg := &AgentConfig{
			ServerAddress:  "baseline:8080",
			ReportInterval: 10 * time.Second,
			PollInterval:   2 * time.Second,
		}

		cfg, err := createConfig(baseCfg)

		assert.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, "baseline:8080", cfg.ServerAddress)
		assert.Equal(t, 10*time.Second, cfg.ReportInterval)
		assert.Equal(t, 2*time.Second, cfg.PollInterval)
	})
}
