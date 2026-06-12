package appkit

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// ServerConfig controls the HTTP server created by NewServer.
type ServerConfig struct {
	Port          int
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration
	HealthHandler http.HandlerFunc // defaults to DefaultHealthHandler
}

// DefaultServerConfig returns sensible defaults for an app server.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

// Server is a thin wrapper around http.Server with graceful shutdown support.
type Server struct {
	cfg    ServerConfig
	server *http.Server
	ln     net.Listener
}

// NewServer creates an app server around the given mux.
func NewServer(cfg ServerConfig, mux *http.ServeMux) *Server {
	if cfg.Port == 0 {
		cfg.Port = DefaultServerConfig().Port
	}

	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = DefaultServerConfig().ReadTimeout
	}

	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = DefaultServerConfig().WriteTimeout
	}

	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = DefaultServerConfig().IdleTimeout
	}

	if cfg.HealthHandler == nil {
		cfg.HealthHandler = DefaultHealthHandler
	}

	mux.HandleFunc("GET /health", cfg.HealthHandler)

	return &Server{
		cfg: cfg,
		server: &http.Server{
			Handler:      mux,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

// Start binds the configured port and serves until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	s.ln = ln

	errCh := make(chan error, 1)

	go func() {
		if err := s.server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("serve: %w", err)

			return
		}

		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

// Shutdown gracefully stops the server using the given context.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	return s.server.Shutdown(ctx)
}

// Addr returns the listener address, or nil if Start has not been called.
func (s *Server) Addr() net.Addr {
	if s.ln == nil {
		return nil
	}

	return s.ln.Addr()
}
