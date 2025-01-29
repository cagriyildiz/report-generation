package store

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
	"report-generation/config"
	"strings"
	"testing"
)

type TestDB struct {
	cfg *config.Config
	DB  *sql.DB
}

func NewTestDB(t *testing.T) *TestDB {
	cfg, err := config.New()
	require.NoError(t, err)

	db, err := NewPostgresDB(cfg)
	require.NoError(t, err)

	return &TestDB{
		cfg: cfg,
		DB:  db,
	}
}

func (db TestDB) Setup(t *testing.T) func(t *testing.T) {
	m, err := migrate.New(
		"file://../migration",
		db.cfg.DBUrl,
	)
	require.NoError(t, err)

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		require.NoError(t, err)
	}

	return db.Teardown
}

func (db TestDB) Teardown(t *testing.T) {
	tables := []string{"users", "refresh_tokens", "reports"}
	_, err := db.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s;", strings.Join(tables, ",")))
	require.NoError(t, err)
}
