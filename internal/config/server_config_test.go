package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetServerConfig_Positive(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()

	testCases := []struct {
		name     string
		args     []string
		envValue string
		wantAddr string
	}{
		{
			name:     "No Env + No Flag -> Default address",
			args:     []string{"cmd"},
			envValue: "",
			wantAddr: ":8080",
		},
		{
			name:     "No Env + Flag -> Use flag value",
			args:     []string{"cmd", "-a", "127.0.0.1:9090"},
			envValue: "",
			wantAddr: "127.0.0.1:9090",
		},
		{
			name:     "No Env + Flag in alternative format -> Use flag value",
			args:     []string{"cmd", "-a=0.0.0.0:3000"},
			envValue: "",
			wantAddr: "0.0.0.0:3000",
		},
		{
			name:     "Env + Flag -> Use env variable",
			args:     []string{"cmd", "-a=0.0.0.0:3000"},
			envValue: "127.0.0.1:9090",
			wantAddr: "127.0.0.1:9090",
		},
		{
			name:     "Env + No Flag -> Use env variable",
			args:     []string{"cmd"},
			envValue: "127.0.0.1:9090",
			wantAddr: "127.0.0.1:9090",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envValue != "" {
				t.Setenv(envAddress, tc.envValue)
			}
			cfg, err := SetServerConfig(tc.args[1:])

			assert.NotNil(t, cfg)
			assert.Nil(t, err)
			assert.Equal(t, tc.wantAddr, cfg.ServerAddress)
		})
	}
}

func TestSetServerConfig_Negative(t *testing.T) {
	t.Run("returns error when an invalid flag is passed", func(t *testing.T) {
		cfg, err := SetServerConfig([]string{"-invalid-flag-structure"})

		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}

func TestNewServerConfig(t *testing.T) {
	wantAddr := "localhost:7070"
	cfg := newServerConfig(wantAddr)

	assert.NotNil(t, cfg)
	assert.Equal(t, wantAddr, cfg.ServerAddress)
}
