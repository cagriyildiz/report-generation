package store

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRefreshTokenStore(t *testing.T) {
	testDB := NewTestDB(t)
	cleanup := testDB.Setup(t)
	defer cleanup(t)

	ctx := context.Background()

	userStore := NewUserStore(testDB.DB)
	user, err := userStore.CreateUser(ctx, "test@test.com", "test")
	require.NoError(t, err)

	jwtToken := createToken(t, user.Id)

	refreshTokenStore := NewRefreshTokenStore(testDB.DB)

	createdToken, err := refreshTokenStore.Create(ctx, jwtToken)
	require.NoError(t, err)

	require.Equal(t, user.Id, createdToken.UserId)

	retrievedToken, err := refreshTokenStore.GetByPrimaryKey(ctx, user.Id, jwtToken)
	require.NoError(t, err)

	require.Equal(t, createdToken, retrievedToken)
	require.Equal(t, jwtToken.Raw, retrievedToken.HashedToken)

	result, err := refreshTokenStore.DeleteUserTokens(ctx, user.Id)
	require.NoError(t, err)
	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), rowsAffected)
}

func createToken(t *testing.T, userId uuid.UUID) *jwt.Token {
	now := time.Now()
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "test",
		Subject:   userId.String(),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
	})

	key := []byte("secret")
	token, err := jwtToken.SignedString(key)
	require.NoError(t, err)
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	require.NoError(t, err)
	return parsedToken
}
