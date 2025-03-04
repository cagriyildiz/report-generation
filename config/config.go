package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerPort           string `env:"SERVER_PORT" envDefault:"5000"`
	ServerHost           string `env:"SERVER_HOST" envDefault:"127.0.0.1"`
	DBName               string `env:"DB_NAME" envDefault:"report-generation"`
	DBHost               string `env:"DB_HOST" envDefault:"127.0.0.1"`
	DBUser               string `env:"DB_USER" envDefault:"root"`
	DBPass               string `env:"DB_PASS" envDefault:"secret"`
	DBPort               string `env:"DB_PORT" envDefault:"5432"`
	DBUrl                string `env:"DB_URL" envDefault:"postgresql://root:secret@127.0.0.1:5432/report-generation?sslmode=disable"`
	JWTSecret            string `env:"JWT_SECRET" envDefault:"secret"`
	AWSS3Bucket          string `env:"AWS_S3_BUCKET" envDefault:"api-reports"`
	AWSSQSQueue          string `env:"AWS_SQS_QUEUE" envDefault:"reports-sqs-queue"`
	LocalstackEndpoint   string `env:"LOCALSTACK_ENDPOINT" envDefault:"http://localhost:4566"`
	LocalstackS3Endpoint string `env:"LOCALSTACK_S3_ENDPOINT" envDefault:"http://s3.localhost.localstack.cloud:4566"`
}

func New() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}
