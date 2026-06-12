package appkit

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

type ServerConfig struct {
	Port          int
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration
	HealthHandler http.HandlerFunc
	RegisterHealth bool
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Port:           8080,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
		RegisterHealth: true,
	}
}

func (cfg *ServerConfig) applyDefaults() {
	def := DefaultServerConfig()

	if cfg.Port == 0 {
		cfg.Port = def.Port
	}

	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = def.ReadTimeout
	}

	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = def.WriteTimeout
	}

	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = def.IdleTimeout
	}

	if cfg.HealthHandler == nil {
		cfg.HealthHandler = DefaultHealthHandler
	}
}

type Server struct {
	cfg    ServerConfig
	server *http.Server
	ln     net.Listener
	mu     sync.RWMutex
}

func NewServer(cfg ServerConfig, mux *http.ServeMux) *Server {
	cfg.applyDefaults()

	if cfg.RegisterHealth {
		mux.HandleFunc("GET /health", cfg.HealthHandler)
	}

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

func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	s.mu.Lock()
	s.ln = ln
	s.mu.Unlock()

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

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	return s.server.Shutdown(ctx)
}

func (s *Server) Addr() net.Addr {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.ln == nil {
		return nil
	}

	return s.ln.Addr()
}

func (s *Server) Running() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.ln != nil
}
