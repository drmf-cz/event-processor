package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/drmf-cz/event-processor/pkg/eventprocessor/nats"
	"go.uber.org/zap"
)

// Server represents the HTTP server for the API
type Server struct {
	cfg     *Config
	logger  *zap.Logger
	handler *Handler
	server  *http.Server
}

// NewServer creates a new API server instance
func NewServer(cfg *Config, logger *zap.Logger, jsClient *nats.JetStreamClient) *Server {
	handler := NewHandler(logger, jsClient)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/messages", handler.PublishMessage)
	mux.HandleFunc("/api/v1/messages/latest", handler.GetLatestMessage)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           mux,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		ReadHeaderTimeout: cfg.ReadTimeout,
	}

	return &Server{
		cfg:     cfg,
		logger:  logger,
		handler: handler,
		server:  server,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", zap.String("addr", s.server.Addr))
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}
	return nil
}

// Stop gracefully shuts down the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server...")
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}
	return nil
}
