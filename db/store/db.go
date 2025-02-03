package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"report-generation/config"

	_ "github.com/lib/pq"
)

func NewPostgresDB(conf *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", conf.DBUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}
	return db, nil
}
