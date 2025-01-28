package store

import (
	"context"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
	"report-generation/config"
	"testing"
)

func TestUserStore(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	db, err := NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	m, err := migrate.New(
		"file://../migration",
		cfg.DBUrl,
	)
	require.NoError(t, err)

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		require.NoError(t, err)
	}

	userStore := NewUserStore(db)
	user, err := userStore.CreateUser(context.Background(), "test@test.com", "test")
	require.NoError(t, err)

	require.Equal(t, "test@test.com", user.Email)
	require.NoError(t, user.ComparePassword("test"))
}
