package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	DBName string `env:"DB_NAME"`
	DBHost string `env:"DB_HOST"`
	DBUser string `env:"DB_USER"`
	DBPass string `env:"DB_PASS"`
	DBPort string `env:"DB_PORT"`
	DBUrl  string `env:"DB_URL"`
}

func New() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}
