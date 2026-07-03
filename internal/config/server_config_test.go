package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetServerConfig(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()

	testCases := []struct {
		name     string
		args     []string
		wantAddr string
	}{
		{
			name:     "uses default address when no flag is provided",
			args:     []string{"cmd"},
			wantAddr: "localhost:8080",
		},
		{
			name:     "overrides address using the custom flag",
			args:     []string{"cmd", "-a", "127.0.0.1:9090"},
			wantAddr: "127.0.0.1:9090",
		},
		{
			name:     "works cleanly with an alternative flag syntax",
			args:     []string{"cmd", "-a=0.0.0.0:3000"},
			wantAddr: "0.0.0.0:3000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			os.Args = tc.args

			cfg := SetServerConfig()

			assert.NotNil(t, cfg)
			assert.Equal(t, tc.wantAddr, cfg.ServerAddress)
		})
	}
}

func TestNewServerConfig(t *testing.T) {
	wantAddr := "localhost:7070"
	cfg := newServerConfig(wantAddr)

	assert.NotNil(t, cfg)
	assert.Equal(t, wantAddr, cfg.ServerAddress)
}
