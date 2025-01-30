package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r SignupRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

type ApiResponse[T any] struct {
	Data    *T     `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

func (s *Server) signupHandler(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	OneMb := int64(1048576)
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, OneMb)).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	existingUser, err := s.store.Users.FindUserByEmail(r.Context(), req.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if existingUser != nil {
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}

	_, err = s.store.Users.CreateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(ApiResponse[struct{}]{
		Message: "successfully signed up user",
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type SigninRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r SigninRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

type SigninResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *Server) signinHandler(w http.ResponseWriter, r *http.Request) {
	var req SigninRequest
	OneMb := int64(1048576)
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, OneMb)).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := s.store.Users.FindUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := user.ComparePassword(req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tokenPair, err := s.jwtManager.CreateTokenPair(user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = s.store.RefreshTokenStore.DeleteUserTokens(r.Context(), user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = s.store.RefreshTokenStore.Create(r.Context(), tokenPair.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).
		Encode(ApiResponse[SigninResponse]{
			Data: &SigninResponse{
				AccessToken:  tokenPair.AccessToken.Raw,
				RefreshToken: tokenPair.RefreshToken.Raw,
			},
		}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type TokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r TokenRefreshRequest) Validate() error {
	if r.RefreshToken == "" {
		return errors.New("refresh token is required")
	}
	return nil
}

type TokenRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *Server) tokenRefreshHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req TokenRefreshRequest
	OneMb := int64(1048576)
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, OneMb)).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jwtToken, err := s.jwtManager.ParseToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	subject, err := jwtToken.Claims.GetSubject()
	if err != nil {
		s.logger.Error("failed to extract subject claim from token", "error", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userId, err := uuid.Parse(subject)
	if err != nil {
		s.logger.Error("token subject is not uuid", "error", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	refreshTokenRecord, err := s.store.RefreshTokenStore.GetByPrimaryKey(ctx, userId, jwtToken)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusUnauthorized
		}
		http.Error(w, "unauthorized", status)
		return
	}

	if refreshTokenRecord.ExpiresAt.Before(time.Now()) {
		http.Error(w, "refresh token is expired", http.StatusUnauthorized)
		return
	}

	tokenPair, err := s.jwtManager.CreateTokenPair(userId)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	_, err = s.store.RefreshTokenStore.DeleteUserTokens(r.Context(), userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = s.store.RefreshTokenStore.Create(r.Context(), tokenPair.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).
		Encode(ApiResponse[TokenRefreshResponse]{
			Data: &TokenRefreshResponse{
				AccessToken:  tokenPair.AccessToken.Raw,
				RefreshToken: tokenPair.RefreshToken.Raw,
			},
		}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
