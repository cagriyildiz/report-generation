package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"report-generation/config"
	"report-generation/db/store"
	"report-generation/reports"
	"time"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatal(err)
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

	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	s3Client := s3.NewFromConfig(awsConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(cfg.LocalstackS3Endpoint)
		options.UsePathStyle = true
	})

	sqsClient := sqs.NewFromConfig(awsConfig, func(options *sqs.Options) {
		options.BaseEndpoint = aws.String(cfg.LocalstackEndpoint)
	})

	lozClient := reports.NewLozClient(&http.Client{
		Timeout: 10 * time.Second,
	})

	reportBuilder := reports.NewReportBuilder(cfg, dataStore.ReportsStore, lozClient, s3Client, logger)

	maxConcurrency := 2
	worker := reports.NewWorker(cfg, reportBuilder, logger, sqsClient, maxConcurrency)

	if err := worker.Start(ctx); err != nil {
		return err
	}

	return nil
}
