package config

import (
	"errors"
	"net"
	"os"
)

const (
	authGrpcHostEnvName = "AUTH_GRPC_HOST"
	authGrpcPortEnvName = "AUTH_GRPC_PORT"
	authCertPathEnvName = "CERT_PATH"
)

type authConfig struct {
	host     string
	port     string
	certPath string
}

// NewAuthConfig returns new auth client config
func NewAuthConfig() (AuthConfig, error) {
	host := os.Getenv(authGrpcHostEnvName)
	if len(host) == 0 {
		return nil, errors.New("auth client host not found")
	}

	port := os.Getenv(authGrpcPortEnvName)
	if len(port) == 0 {
		return nil, errors.New("auth client port not found")
	}

	certPath := os.Getenv(authCertPathEnvName)
	if len(certPath) == 0 {
		return nil, errors.New("cert path not found")
	}

	return &authConfig{
		host:     host,
		port:     port,
		certPath: certPath,
	}, nil
}

// Address returns a full address of a auth client
func (cfg *authConfig) Address() string {
	return net.JoinHostPort(cfg.host, cfg.port)
}

func (cfg *authConfig) CertPath() string {
	return cfg.certPath
}
