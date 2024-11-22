package config

import (
	"github.com/joho/godotenv"
)

// GRPCConfig is interface of a grpc config
type GRPCConfig interface {
	Address() string
}

// PGConfig is interface of a postgres config
type PGConfig interface {
	DSN() string
}

// HTTPConfig интерфейс конфига http-сервера
type HTTPConfig interface {
	Address() string
}

// SwaggerConfig интерфейс конфига swagger http-сервера
type SwaggerConfig interface {
	Address() string
}

// AuthConfig интерфейс конфига клиента auth
type AuthConfig interface {
	Address() string
	CertPath() string
}

// Load reads ,env file from path and loads
// variables into a project
func Load(path string) error {
	err := godotenv.Load(path)
	if err != nil {
		return err
	}

	return nil
}
