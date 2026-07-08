package config

import (
	"flag"
	"os"
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

	// Rewrite config with env variable if any
	if envAddr := getEnvAddrVariable(); envAddr != "" {
		address = &envAddr
	}
	return newServerConfig(*address), nil
}

func getEnvAddrVariable() string {
	return os.Getenv(envAddress)
}
