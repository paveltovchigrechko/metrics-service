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

func SetServerConfig() *ServerConfig {
	address := flag.String(serverAddressFlag, defaultAddress, "Metrics server HTTP address")
	flag.Parse()
	return newServerConfig(*address)
}
