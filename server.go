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

const (
	defaultPort         = 8080
	defaultReadTimeout  = 30 * time.Second
	defaultWriteTimeout = 30 * time.Second
	defaultIdleTimeout  = 120 * time.Second
	defaultKeepAlive    = 30 * time.Second
)

type ServerConfig struct {
	Port           int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	HealthHandler  http.HandlerFunc
	RegisterHealth bool
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Port:           defaultPort,
		ReadTimeout:    defaultReadTimeout,
		WriteTimeout:   defaultWriteTimeout,
		IdleTimeout:    defaultIdleTimeout,
		HealthHandler:  DefaultHealthHandler,
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
		ln:  nil,
		mu:  sync.RWMutex{},
		server: &http.Server{
			Addr:                         "",
			Handler:                      mux,
			ReadTimeout:                  cfg.ReadTimeout,
			WriteTimeout:                 cfg.WriteTimeout,
			IdleTimeout:                  cfg.IdleTimeout,
			ReadHeaderTimeout:            cfg.ReadTimeout,
			DisableGeneralOptionsHandler: false,
			TLSConfig:                    nil,
			MaxHeaderBytes:               http.DefaultMaxHeaderBytes,
			TLSNextProto:                 nil,
			ConnState:                    nil,
			ErrorLog:                     nil,
			BaseContext:                  nil,
			ConnContext:                  nil,
			HTTP2:                        nil,
			Protocols:                    nil,
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)

	listenConfig := &net.ListenConfig{
		Control:         nil,
		KeepAlive:       defaultKeepAlive,
		KeepAliveConfig: net.KeepAliveConfig{Enable: false, Idle: 0, Interval: 0, Count: 0},
	}

	listener, err := listenConfig.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	s.mu.Lock()
	s.ln = listener
	s.mu.Unlock()

	errCh := make(chan error, 1)

	go func() {
		err := s.server.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
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

	err := s.server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	return nil
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
