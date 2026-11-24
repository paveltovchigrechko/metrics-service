package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(path, []byte(body), 0600))
	return path
}

// clearConfigEnv unsets every variable the config loader reads, so a file
// test is not affected by the developer's environment.
func clearConfigEnv(t *testing.T) {
	t.Helper()

	for _, name := range []string{
		"ADDRESS",
		"REPORT_INTERVAL",
		"POLL_INTERVAL",
		"STORE_INTERVAL",
		"FILE_STORAGE_PATH",
		"RESTORE",
		"DATABASE_DSN",
		"KEY",
		"RATE_LIMIT",
		"AUDIT_FILE",
		"AUDIT_URL",
		"CRYPTO_KEY",
		"TRUSTED_SUBNET",
		"CONFIG",
	} {
		t.Setenv(name, "")
	}
}

func TestSetAgentConfig_File(t *testing.T) {
	path := writeConfig(t, `{
		"address": "10.0.0.8:9090",
		"report_interval": "1m",
		"poll_interval": "3s",
		"key": "file-key",
		"rate_limit": 7,
		"crypto_key": " /file/public.pem "
	}`)

	t.Run("file fills every option", func(t *testing.T) {
		clearConfigEnv(t)

		cfg, err := SetAgentConfig([]string{"-c", path})
		require.NoError(t, err)

		assert.Equal(t, "10.0.0.8:9090", cfg.ServerAddress)
		assert.Equal(t, time.Minute, cfg.ReportInterval)
		assert.Equal(t, 3*time.Second, cfg.PollInterval)
		assert.Equal(t, "file-key", cfg.Key)
		assert.Equal(t, 7, cfg.RateLimit)
		assert.Equal(t, "/file/public.pem", cfg.CryptoKey)
	})

	t.Run("long flag is an alias of -c", func(t *testing.T) {
		clearConfigEnv(t)

		cfg, err := SetAgentConfig([]string{"--config", path})
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.8:9090", cfg.ServerAddress)
		assert.Equal(t, 7, cfg.RateLimit)
	})

	t.Run("omitted keys keep defaults", func(t *testing.T) {
		clearConfigEnv(t)
		onlyAddress := writeConfig(t, `{"address": "10.0.0.9:1"}`)

		cfg, err := SetAgentConfig([]string{"-config", onlyAddress})
		require.NoError(t, err)

		assert.Equal(t, "10.0.0.9:1", cfg.ServerAddress)
		assert.Equal(t, 10*time.Second, cfg.ReportInterval)
		assert.Equal(t, 2*time.Second, cfg.PollInterval)
		assert.Empty(t, cfg.Key)
		assert.Equal(t, 2, cfg.RateLimit)
		assert.Empty(t, cfg.CryptoKey)
	})

	t.Run("empty object keeps defaults", func(t *testing.T) {
		clearConfigEnv(t)

		cfg, err := SetAgentConfig([]string{"-c", writeConfig(t, `{}`)})
		require.NoError(t, err)

		assert.Equal(t, defaultServerAddress, cfg.ServerAddress)
		assert.Equal(t, defaultReportInterval*time.Second, cfg.ReportInterval)
		assert.Equal(t, defaultRateLimit, cfg.RateLimit)
	})
}

func TestSetAgentConfig_FilePriority(t *testing.T) {
	filePath := writeConfig(t, `{
		"address": "file:1",
		"poll_interval": "3s",
		"rate_limit": 7,
		"crypto_key": "/file/public.pem"
	}`)
	otherPath := writeConfig(t, `{"address": "other:1", "rate_limit": 4}`)

	testCases := []struct {
		name     string
		args     []string
		env      map[string]string
		wantAddr string
		wantPoll time.Duration
		wantRate int
		wantKey  string
	}{
		{
			name:     "flag overrides the file and leaves other file values",
			args:     []string{"-c", filePath, "-a", "flag:1", "-p", "8"},
			wantAddr: "flag:1",
			wantPoll: 8 * time.Second,
			wantRate: 7,
			wantKey:  "/file/public.pem",
		},
		{
			name:     "explicit flag equal to the default beats the file",
			args:     []string{"-c", filePath, "-a", defaultServerAddress},
			wantAddr: defaultServerAddress,
			wantPoll: 3 * time.Second,
			wantRate: 7,
			wantKey:  "/file/public.pem",
		},
		{
			name:     "environment overrides the file and the flag",
			args:     []string{"-c", filePath, "-a", "flag:1", "-l", "10"},
			env:      map[string]string{"ADDRESS": "env:1", "POLL_INTERVAL": "9", "RATE_LIMIT": "5"},
			wantAddr: "env:1",
			wantPoll: 9 * time.Second,
			wantRate: 5,
			wantKey:  "/file/public.pem",
		},
		{
			name:     "CONFIG selects the file over -c",
			args:     []string{"-c", otherPath},
			env:      map[string]string{"CONFIG": filePath},
			wantAddr: "file:1",
			wantPoll: 3 * time.Second,
			wantRate: 7,
			wantKey:  "/file/public.pem",
		},
		{
			name:     "empty explicit crypto-key flag beats the file",
			args:     []string{"-c", filePath, "-crypto-key="},
			wantAddr: "file:1",
			wantPoll: 3 * time.Second,
			wantRate: 7,
			wantKey:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clearConfigEnv(t)
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			cfg, err := SetAgentConfig(tc.args)
			require.NoError(t, err)
			assert.Equal(t, tc.wantAddr, cfg.ServerAddress)
			assert.Equal(t, tc.wantPoll, cfg.PollInterval)
			assert.Equal(t, tc.wantRate, cfg.RateLimit)
			assert.Equal(t, tc.wantKey, cfg.CryptoKey)
		})
	}
}

func TestSetAgentConfig_FileErrors(t *testing.T) {
	clearConfigEnv(t)

	t.Run("missing file", func(t *testing.T) {
		cfg, err := SetAgentConfig([]string{"-c", filepath.Join(t.TempDir(), "missing.json")})
		require.ErrorIs(t, err, os.ErrNotExist)
		assert.Nil(t, cfg)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := SetAgentConfig([]string{"-c", writeConfig(t, `{`)})
		assert.Error(t, err)
	})

	t.Run("unknown field", func(t *testing.T) {
		_, err := SetAgentConfig([]string{"-c", writeConfig(t, `{"nope": 1}`)})
		assert.Error(t, err)
	})

	t.Run("duration without a unit", func(t *testing.T) {
		_, err := SetAgentConfig([]string{"-c", writeConfig(t, `{"report_interval": "15"}`)})
		assert.Error(t, err)
	})

	t.Run("non-positive interval", func(t *testing.T) {
		_, err := SetAgentConfig([]string{"-c", writeConfig(t, `{"report_interval": "0s"}`)})
		assert.ErrorIs(t, err, errIncorrectInterval)
	})

	t.Run("bad duration fails even when the flag overrides it", func(t *testing.T) {
		_, err := SetAgentConfig([]string{"-c", writeConfig(t, `{"poll_interval": "nope"}`), "-p", "4"})
		assert.Error(t, err)
	})
}

func TestSetServerConfig_File(t *testing.T) {
	path := writeConfig(t, `{
		"address": "10.1.1.1:8081",
		"restore": false,
		"store_interval": "45s",
		"store_file": "/tmp/metrics.db",
		"database_dsn": "postgres://localhost/db",
		"key": "hmac",
		"audit_file": "/tmp/audit.log",
		"audit_url": "http://audit",
		"crypto_key": " /tmp/private.pem "
	}`)

	t.Run("file fills every option", func(t *testing.T) {
		clearConfigEnv(t)

		cfg, err := SetServerConfig([]string{"--config", path})
		require.NoError(t, err)

		assert.Equal(t, "10.1.1.1:8081", cfg.ServerAddress)
		assert.False(t, cfg.Restore)
		assert.Equal(t, 45*time.Second, cfg.StoreInterval)
		assert.Equal(t, "/tmp/metrics.db", cfg.FileStoragePath)
		assert.Equal(t, "postgres://localhost/db", cfg.DatabaseDSN)
		assert.Equal(t, "hmac", cfg.Key)
		assert.Equal(t, "/tmp/audit.log", cfg.AuditFilePath)
		assert.Equal(t, "http://audit", cfg.AuditURL)
		assert.Equal(t, "/tmp/private.pem", cfg.CryptoKey)
	})

	t.Run("omitted restore keeps the default", func(t *testing.T) {
		clearConfigEnv(t)

		cfg, err := SetServerConfig([]string{"-c", writeConfig(t, `{"address": "10.2.2.2:1"}`)})
		require.NoError(t, err)

		assert.Equal(t, "10.2.2.2:1", cfg.ServerAddress)
		assert.True(t, cfg.Restore)
		assert.Equal(t, defaultStoreInterval*time.Second, cfg.StoreInterval)
		assert.Equal(t, defaultFileStoragePath, cfg.FileStoragePath)
	})

	t.Run("zero store interval is kept", func(t *testing.T) {
		clearConfigEnv(t)

		cfg, err := SetServerConfig([]string{"-c", writeConfig(t, `{"store_interval": "0s"}`)})
		require.NoError(t, err)
		assert.Equal(t, time.Duration(0), cfg.StoreInterval)
		assert.True(t, cfg.Restore)
	})
}

func TestSetServerConfig_FilePriority(t *testing.T) {
	filePath := writeConfig(t, `{
		"address": "file:1",
		"restore": false,
		"store_interval": "45s",
		"store_file": "/file.db",
		"crypto_key": "/file/private.pem"
	}`)

	t.Run("flag and environment override the file", func(t *testing.T) {
		clearConfigEnv(t)
		t.Setenv("FILE_STORAGE_PATH", "/env.db")
		t.Setenv("RESTORE", "true")

		cfg, err := SetServerConfig([]string{"-c", filePath, "-a", "flag:1", "-i", "10"})
		require.NoError(t, err)

		assert.Equal(t, "flag:1", cfg.ServerAddress)
		assert.Equal(t, 10*time.Second, cfg.StoreInterval)
		assert.Equal(t, "/env.db", cfg.FileStoragePath)
		assert.True(t, cfg.Restore)
		assert.Equal(t, "/file/private.pem", cfg.CryptoKey)
	})

	t.Run("CONFIG selects the file over -c", func(t *testing.T) {
		clearConfigEnv(t)
		t.Setenv("CONFIG", filePath)

		other := writeConfig(t, `{"address": "other:1", "restore": true}`)
		cfg, err := SetServerConfig([]string{"-c", other})
		require.NoError(t, err)

		assert.Equal(t, "file:1", cfg.ServerAddress)
		assert.False(t, cfg.Restore)
	})

	t.Run("negative interval", func(t *testing.T) {
		clearConfigEnv(t)

		_, err := SetServerConfig([]string{"-c", writeConfig(t, `{"store_interval": "-1s"}`)})
		assert.ErrorIs(t, err, errIncorrectInterval)
	})

	t.Run("unknown field", func(t *testing.T) {
		clearConfigEnv(t)

		_, err := SetServerConfig([]string{"-config", writeConfig(t, `{"store_path": "/x"}`)})
		assert.Error(t, err)
	})
}
