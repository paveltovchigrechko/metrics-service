package config

import "flag"

type ServerConfig struct {
	ServerAddress string
}

const (
	serverAddressFlag = "a"
	defaultAddress    = "localhost:8080"
)

func newServerConfig(addr string) *ServerConfig {
	return &ServerConfig{
		ServerAddress: addr,
	}
}

func SetServerConfig(args []string) (*ServerConfig, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	address := fs.String(serverAddressFlag, defaultAddress, "Metrics server HTTP address")
	err := fs.Parse(args)

	if err != nil {
		return nil, err
	}
	return newServerConfig(*address), nil
}
