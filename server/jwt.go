package server

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"report-generation/config"
)

type JwtManager struct {
	cfg *config.Config
}

func NewJwtManager(cfg *config.Config) *JwtManager {
	return &JwtManager{
		cfg: cfg,
	}
}

type TokenPair struct {
	AccessToken  *jwt.Token
	RefreshToken *jwt.Token
}

type CustomClaims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func (m *JwtManager) CreateTokenPair(userId uuid.UUID) (*TokenPair, error) {
	accessToken, err := m.createToken(userId, "access")
	if err != nil {
		return nil, err
	}

	refreshToken, err := m.createToken(userId, "refresh")
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (m *JwtManager) ParseToken(token string) (*jwt.Token, error) {
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}
	jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(m.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}
	return jwtToken, nil
}

func (m *JwtManager) IsAccessToken(token *jwt.Token) bool {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	if tokenType, ok := claims["token_type"]; ok {
		return tokenType == "access"
	}
	return false
}

func (m *JwtManager) createToken(userId uuid.UUID, tokenType string) (*jwt.Token, error) {
	var expiration = time.Minute * 15
	if tokenType == "refresh" {
		expiration = time.Hour * 24
	}
	now := time.Now()
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256,
		CustomClaims{
			TokenType: tokenType,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    fmt.Sprintf("https://%s:%s", m.cfg.ServerHost, m.cfg.ServerPort),
				Subject:   userId.String(),
				ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
				IssuedAt:  jwt.NewNumericDate(now),
			},
		},
	)
	key := []byte(m.cfg.JWTSecret)
	signed, err := jwtToken.SignedString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %v", err)
	}
	return m.ParseToken(signed)
}
