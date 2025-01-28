package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"report-generation/config"
	"report-generation/server"
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
	srv := server.New(cfg, logger)
	if err := srv.Start(ctx); err != nil {
		return err
	}
	return nil
}
