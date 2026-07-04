package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetAgentConfig(t *testing.T) {
	testCases := []struct {
		name       string
		args       []string
		wantAddr   string
		wantReport time.Duration
		wantPoll   time.Duration
		wantErr    error
	}{
		{
			name:       "uses default config parameters when no flags are provided",
			args:       []string{},
			wantAddr:   "localhost:8080",
			wantReport: 10 * time.Second,
			wantPoll:   2 * time.Second,
			wantErr:    nil,
		},
		{
			name:       "overrides all config parameters cleanly via custom flags",
			args:       []string{"-a", "127.0.0.1:9090", "-r", "30", "-p", "5"},
			wantAddr:   "127.0.0.1:9090",
			wantReport: 30 * time.Second,
			wantPoll:   5 * time.Second,
			wantErr:    nil,
		},
		{
			name:       "fails validation if report interval is zero or negative",
			args:       []string{"-r", "0"},
			wantAddr:   "",
			wantReport: 0,
			wantPoll:   0,
			wantErr:    errIncorrectInterval,
		},
		{
			name:       "fails validation if poll interval is zero or negative",
			args:       []string{"-p", "-5"},
			wantAddr:   "",
			wantReport: 0,
			wantPoll:   0,
			wantErr:    errIncorrectInterval,
		},
		{
			name:       "returns error when an unknown flag is specified",
			args:       []string{"-unknown-flag"},
			wantAddr:   "",
			wantReport: 0,
			wantPoll:   0,
			wantErr:    assert.AnError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := SetAgentConfig(tc.args)

			if tc.wantErr != nil {
				assert.Error(t, err)

				if tc.wantErr == errIncorrectInterval {
					assert.ErrorIs(t, err, errIncorrectInterval)
				}
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				assert.Equal(t, tc.wantAddr, cfg.ServerAddress)
				assert.Equal(t, tc.wantReport, cfg.ReportInterval)
				assert.Equal(t, tc.wantPoll, cfg.PollInterval)
			}
		})
	}
}

func TestValidateIntervals(t *testing.T) {
	testCases := []struct {
		name        string
		reportInt   int
		pollInt     int
		expectedErr error
	}{
		{
			name:        "both intervals positive is valid",
			reportInt:   10,
			pollInt:     2,
			expectedErr: nil,
		},
		{
			name:        "zero report interval is invalid",
			reportInt:   0,
			pollInt:     2,
			expectedErr: errIncorrectInterval,
		},
		{
			name:        "negative poll interval is invalid",
			reportInt:   10,
			pollInt:     -1,
			expectedErr: errIncorrectInterval,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateIntervals(tc.reportInt, tc.pollInt)
			if tc.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.expectedErr)
			}
		})
	}
}
