package server

import (
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"report-generation/db/store"
	"strings"
)

func NewLoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("http request", "method", r.Method, "path", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}

func NewAuthMiddleware(logger *slog.Logger, jwtManager *JwtManager, userStore *store.UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.RequestURI, "/auth") {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			header := r.Header.Get("Authorization")
			parts := strings.Split(header, "Bearer ")

			if len(parts) != 2 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			token := parts[1]
			jwtToken, err := jwtManager.ParseToken(token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if !jwtManager.IsAccessToken(jwtToken) {
				http.Error(w, "not an access token", http.StatusUnauthorized)
				return
			}

			subject, err := jwtToken.Claims.GetSubject()
			if err != nil {
				logger.Error("failed to extract subject claim from token", "error", err)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userId, err := uuid.Parse(subject)
			if err != nil {
				logger.Error("token subject is not uuid", "error", err)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := userStore.FindUserById(ctx, userId)
			if err != nil {
				logger.Error("failed to find user", "error", err)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ContextWithUser(ctx, user)))
		})
	}
}
