package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"log/slog"
	"os"
	"os/signal"
	"report-generation/config"
	"report-generation/db/store"
	"report-generation/server"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		return err
	}
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	)
	db, err := store.NewPostgresDB(cfg)
	if err != nil {
		return err
	}
	dataStore := store.New(db)
	jwtManager := server.NewJwtManager(cfg)

	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("couldn't load default configuration", "error", err)
		return err
	}

	sqsClient := sqs.NewFromConfig(awsConfig, func(options *sqs.Options) {
		options.BaseEndpoint = aws.String(cfg.LocalstackEndpoint)
	})

	srv := server.New(cfg, logger, dataStore, jwtManager, sqsClient)
	if err := srv.Start(ctx); err != nil {
		return err
	}
	return nil
}
