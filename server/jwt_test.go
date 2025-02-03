package server

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"report-generation/config"
	"testing"
)

func TestJwtManager(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	jwtManager := NewJwtManager(cfg)
	userId := uuid.New()
	tokenPair, err := jwtManager.CreateTokenPair(userId)
	require.NoError(t, err)

	accessToken, err := jwtManager.ParseToken(tokenPair.AccessToken.Raw)
	require.NoError(t, err)

	require.True(t, jwtManager.IsAccessToken(accessToken))

	require.Equal(t, tokenPair.AccessToken.Raw, accessToken.Raw)

	accessTokenSubject, err := accessToken.Claims.GetSubject()
	require.NoError(t, err)
	require.Equal(t, userId.String(), accessTokenSubject)

	accessTokenIssuer, err := accessToken.Claims.GetIssuer()
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("https://%s:%s", cfg.ServerHost, cfg.ServerPort), accessTokenIssuer)

	// ---

	refreshToken, err := jwtManager.ParseToken(tokenPair.RefreshToken.Raw)
	require.NoError(t, err)

	require.False(t, jwtManager.IsAccessToken(refreshToken))

	require.Equal(t, tokenPair.RefreshToken.Raw, refreshToken.Raw)

	refreshTokenSubject, err := refreshToken.Claims.GetSubject()
	require.NoError(t, err)
	require.Equal(t, userId.String(), refreshTokenSubject)

	refreshTokenIssuer, err := refreshToken.Claims.GetIssuer()
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("https://%s:%s", cfg.ServerHost, cfg.ServerPort), refreshTokenIssuer)
}
