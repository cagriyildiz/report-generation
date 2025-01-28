package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"report-generation/config"
	"sync"
	"time"
)

type Server struct {
	cfg    *config.Config
	logger *slog.Logger
}

func New(cfg *config.Config, logger *slog.Logger) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", s.ping)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(s.cfg.ServerHost, s.cfg.ServerPort),
		Handler: mux,
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
