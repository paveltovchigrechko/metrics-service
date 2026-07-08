package config

import (
	"flag"
	"os"
	"strings"
)

type ServerConfig struct {
	ServerAddress string
}

const (
	serverAddressFlag = "a"
	defaultAddress    = "localhost:8080"
	envAddress        = "ADDRESS"
)

func newServerConfig(addr string) *ServerConfig {
	return &ServerConfig{
		ServerAddress: addr,
	}
}

func SetServerConfig(args []string) (*ServerConfig, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	address := fs.String(serverAddressFlag, defaultAddress, "Metrics server HTTP address and port")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	addr := *address

	// Rewrite config with env variable if any
	if envAddr := getEnvAddrVariable(); envAddr != "" {
		addr = envAddr
	}

	if strings.HasPrefix(addr, "localhost:") {
		addr = strings.TrimPrefix(addr, "localhost")
	}
	return newServerConfig(addr), nil
}

func getEnvAddrVariable() string {
	return os.Getenv(envAddress)
}
