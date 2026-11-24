package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetServerConfig_TrustedSubnet(t *testing.T) {
	filePath := writeConfig(t, `{"trusted_subnet": " 10.0.0.0/8 "}`)

	testCases := []struct {
		name string
		args []string
		env  map[string]string
		want string
	}{
		{name: "empty by default", want: ""},
		{name: "-t flag", args: []string{"-t", "192.168.0.0/16"}, want: "192.168.0.0/16"},
		{name: "-t flag is trimmed", args: []string{"-t", " 192.168.0.0/16 "}, want: "192.168.0.0/16"},
		{name: "environment variable", env: map[string]string{"TRUSTED_SUBNET": "172.16.0.0/12"}, want: "172.16.0.0/12"},
		{name: "environment beats the flag", args: []string{"-t", "192.168.0.0/16"}, env: map[string]string{"TRUSTED_SUBNET": "172.16.0.0/12"}, want: "172.16.0.0/12"},
		{name: "config file", args: []string{"-c", filePath}, want: "10.0.0.0/8"},
		{name: "flag beats the config file", args: []string{"-c", filePath, "-t", "192.168.0.0/16"}, want: "192.168.0.0/16"},
		{name: "environment beats the config file", args: []string{"-c", filePath}, env: map[string]string{"TRUSTED_SUBNET": "172.16.0.0/12"}, want: "172.16.0.0/12"},
		{name: "empty flag beats the config file", args: []string{"-c", filePath, "-t="}, want: ""},
		{name: "IPv6 subnet", args: []string{"-t", "2001:db8::/32"}, want: "2001:db8::/32"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clearConfigEnv(t)
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			cfg, err := SetServerConfig(tc.args)
			require.NoError(t, err)
			assert.Equal(t, tc.want, cfg.TrustedSubnet)
		})
	}
}

func TestSetServerConfig_TrustedSubnetInvalid(t *testing.T) {
	t.Run("flag", func(t *testing.T) {
		clearConfigEnv(t)

		cfg, err := SetServerConfig([]string{"-t", "192.168.1.1"})
		require.ErrorIs(t, err, errInvalidTrustedSubnet)
		assert.Nil(t, cfg)
	})

	t.Run("environment", func(t *testing.T) {
		clearConfigEnv(t)
		t.Setenv("TRUSTED_SUBNET", "10.0.0.0/40")

		_, err := SetServerConfig(nil)
		assert.ErrorIs(t, err, errInvalidTrustedSubnet)
	})

	t.Run("config file", func(t *testing.T) {
		clearConfigEnv(t)

		_, err := SetServerConfig([]string{"-c", writeConfig(t, `{"trusted_subnet": "nope"}`)})
		assert.ErrorIs(t, err, errInvalidTrustedSubnet)
	})
}

func TestCreateServerFlagConfig_TrustedSubnet(t *testing.T) {
	cfg, explicit, _, err := createServerFlagConfig([]string{"-t", "10.1.0.0/16"})

	require.NoError(t, err)
	assert.Equal(t, "10.1.0.0/16", cfg.TrustedSubnet)
	assert.True(t, explicit[trustedSubnetFlag])
}

func TestParseEnvServerConfig_TrustedSubnet(t *testing.T) {
	t.Run("set", func(t *testing.T) {
		clearConfigEnv(t)
		t.Setenv("TRUSTED_SUBNET", "10.1.0.0/16")

		envCfg, err := parseEnvServerConfig()
		require.NoError(t, err)
		require.NotNil(t, envCfg.TrustedSubnet)
		assert.Equal(t, "10.1.0.0/16", *envCfg.TrustedSubnet)
	})

	t.Run("unset", func(t *testing.T) {
		clearConfigEnv(t)

		envCfg, err := parseEnvServerConfig()
		require.NoError(t, err)
		assert.Nil(t, envCfg.TrustedSubnet)
	})
}
