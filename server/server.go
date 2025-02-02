package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"report-generation/config"
	"report-generation/db/store"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Server struct {
	cfg        *config.Config
	logger     *slog.Logger
	store      *store.Store
	jwtManager *JwtManager
	sqsClient  *sqs.Client
}

func New(
	cfg *config.Config,
	logger *slog.Logger,
	store *store.Store,
	jwtManager *JwtManager,
	sqsClient *sqs.Client,
) *Server {
	return &Server{
		cfg:        cfg,
		logger:     logger,
		store:      store,
		jwtManager: jwtManager,
		sqsClient:  sqsClient,
	}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", s.ping)
	mux.HandleFunc("POST /auth/signup", s.signupHandler)
	mux.HandleFunc("POST /auth/signin", s.signinHandler)
	mux.HandleFunc("POST /auth/refresh", s.tokenRefreshHandler)
	mux.HandleFunc("POST /reports", s.createReportHandler)
	mux.HandleFunc("GET /reports/{id}", s.getReportHandler)

	middleware := NewLoggerMiddleware(s.logger)
	middleware = NewAuthMiddleware(s.logger, s.jwtManager, s.store.Users)

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(s.cfg.ServerHost, s.cfg.ServerPort),
		Handler: middleware(mux),
	}

	go func() {
		s.logger.Info("server is running", "port", s.cfg.ServerPort)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("server failed to listen and serve", "error", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("server failed to shutdown", "error", err)
		}
	}()
	wg.Wait()

	return nil
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}
