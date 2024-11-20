package config

import (
	"errors"
	"net"
	"os"
)

const (
	authGrpcHostEnvName = "AUTH_GRPC_HOST"
	authGrpcPortEnvName = "AUTH_GRPC_PORT"
)

type authConfig struct {
	host string
	port string
}

// NewAuthonfig returns new auth client config
func NewAuthConfig() (AuthConfig, error) {
	host := os.Getenv(authGrpcHostEnvName)
	if len(host) == 0 {
		return nil, errors.New("auth client host not found")
	}

	port := os.Getenv(authGrpcPortEnvName)
	if len(port) == 0 {
		return nil, errors.New("auth client port not found")
	}

	return &grpcConfig{
		host: host,
		port: port,
	}, nil
}

// Address returns a full address of a auth client
func (cfg *authConfig) Address() string {
	return net.JoinHostPort(cfg.host, cfg.port)
}
