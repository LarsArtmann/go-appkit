package appkit

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/httputil"
)

// Service is a production-ready HTTP service with middleware, health checks,
// structured logging, and graceful shutdown.
type Service struct {
	cfg ServiceConfig

	Logger *slog.Logger
	Mux    *http.ServeMux

	server     *http.Server
	ln         net.Listener
	readyProbe atomic.Bool
	mu         sync.RWMutex
}

// NewService creates a Service with the given configuration.
// Zero-value fields in cfg are replaced with production defaults.
func NewService(cfg ServiceConfig) (*Service, error) {
	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	logger, err := InitLogger(LoggerConfig{Level: cfg.LogLevel, Format: cfg.LogFormat})
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	svc := &Service{
		cfg:    cfg,
		Logger: logger,
		Mux:    mux,
	}

	svc.readyProbe.Store(true)

	if cfg.RegisterHealth == nil || *cfg.RegisterHealth {
		mux.HandleFunc("GET /health", httputil.HealthHandler())
		mux.HandleFunc("GET /health/live", httputil.LiveHandler())
		mux.HandleFunc("GET /health/ready", httputil.ReadyHandlerWithProbe(svc.ready))
	}

	mws := buildMiddleware(logger, cfg)
	wrapped := httputil.Chain(mux, mws...)

	svc.server = &http.Server{
		Handler:           wrapped,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	return svc, nil
}

// Start creates the listener and begins serving in a goroutine.
// Returns a synchronous error if the listener cannot bind.
// The returned channel receives any serve error (nil on graceful shutdown).
func (s *Service) Start() (<-chan error, error) {
	ln, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "server.listen_failed", "listen on %s", s.cfg.Addr)
	}

	s.mu.Lock()
	s.ln = ln
	s.mu.Unlock()

	errCh := make(chan error, 1)

	go func() {
		serveErr := s.server.Serve(ln)
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- errorfamily.WrapInfrastructuref(serveErr, "server.serve_failed", "serve failed")

			return
		}

		errCh <- nil
	}()

	return errCh, nil
}

// Run starts the service and blocks until a signal is received, ctx is cancelled,
// or the server encounters an error. Handles graceful drain + shutdown internally.
func (s *Service) Run(ctx context.Context) error {
	errCh, err := s.Start()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		s.Logger.Info("context cancelled, shutting down")
	case serveErr := <-errCh:
		return serveErr
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	return s.Shutdown(shutdownCtx)
}

// Shutdown performs the graceful drain sequence:
// 1. Flip ready probe to false (load balancer stops sending traffic)
// 2. Wait DrainDelay for load balancer propagation
// 3. Stop accepting new connections and finish in-flight requests.
func (s *Service) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	ln := s.ln
	s.ln = nil
	s.mu.Unlock()

	if ln == nil {
		return nil
	}

	s.readyProbe.Store(false)

	if s.cfg.DrainDelay > 0 {
		s.Logger.Info("draining traffic", "delay", s.cfg.DrainDelay)

		timer := time.NewTimer(s.cfg.DrainDelay)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-ctx.Done():
		}
	}

	err := s.server.Shutdown(ctx)
	if err != nil {
		return errorfamily.WrapInfrastructuref(err, "server.shutdown_failed", "shutdown failed")
	}

	return nil
}

// Close calls Shutdown with context.Background(). Idempotent.
func (s *Service) Close() error {
	return s.Shutdown(context.Background())
}

// Addr returns the bound listener address. Returns nil before Start is called.
func (s *Service) Addr() net.Addr {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.ln == nil {
		return nil
	}

	return s.ln.Addr()
}

// Running reports whether the service has a bound listener.
func (s *Service) Running() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.ln != nil
}

// ready composes the internal drain probe with the optional configured
// ReadyCheck: the service is ready only while the probe is up AND the
// external check (if any) passes. Drain therefore always forces 503.
func (s *Service) ready() bool {
	if !s.readyProbe.Load() {
		return false
	}

	return s.cfg.ReadyCheck == nil || s.cfg.ReadyCheck()
}
