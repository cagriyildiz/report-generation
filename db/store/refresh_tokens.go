package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RefreshTokenStore struct {
	db *sqlx.DB
}

func NewRefreshTokenStore(db *sql.DB) *RefreshTokenStore {
	return &RefreshTokenStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type RefreshToken struct {
	UserId      uuid.UUID `db:"user_id"`
	HashedToken string    `db:"hashed_token"`
	CreatedAt   time.Time `db:"created_at"`
	ExpiresAt   time.Time `db:"expires_at"`
}

func (s *RefreshTokenStore) Create(ctx context.Context, token *jwt.Token) (*RefreshToken, error) {
	const query = `INSERT INTO refresh_tokens (user_id, hashed_token, expires_at) VALUES ($1, $2, $3) RETURNING *`

	var refreshToken RefreshToken
	userId, err := token.Claims.GetSubject()
	if err != nil {
		return nil, fmt.Errorf("failed to extract user id from token: %w", err)
	}

	expiresAt, err := token.Claims.GetExpirationTime()
	if err != nil {
		return nil, fmt.Errorf("failed to extract expiration time from token: %w", err)
	}
	err = s.db.GetContext(ctx, &refreshToken, query, userId, token.Raw, expiresAt.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to insert refresh token: %w", err)
	}

	return &refreshToken, nil
}

func (s *RefreshTokenStore) GetByPrimaryKey(ctx context.Context, userId uuid.UUID, token *jwt.Token) (*RefreshToken, error) {
	const query = "SELECT * FROM refresh_tokens WHERE user_id = $1 AND hashed_token = $2"

	var refreshToken RefreshToken
	err := s.db.GetContext(ctx, &refreshToken, query, userId, token.Raw)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch refresh token: %w", err)
	}

	return &refreshToken, nil
}

func (s *RefreshTokenStore) DeleteUserTokens(ctx context.Context, userId uuid.UUID) (sql.Result, error) {
	const query = "DELETE FROM refresh_tokens WHERE user_id = $1"

	result, err := s.db.ExecContext(ctx, query, userId)
	if err != nil {
		return result, fmt.Errorf("failed to delete refresh tokens: %w", err)
	}
	return result, nil
}
