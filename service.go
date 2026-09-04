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

	err := cfg.Validate()
	if err != nil {
		return nil, err
	}

	logger, err := InitLogger(LoggerConfig{Level: cfg.LogLevel, Format: cfg.LogFormat})
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	svc := &Service{ //nolint:exhaustruct_v5 // server, ln, readyProbe, mu are deliberate zero values
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
		ReadTimeout:       serverTimeout(cfg.ReadTimeout),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      serverTimeout(cfg.WriteTimeout),
		IdleTimeout:       cfg.IdleTimeout,
	}

	return svc, nil
}

// Start creates the listener and begins serving in a goroutine.
// Returns a synchronous error if the listener cannot bind.
// The returned channel receives any serve error (nil on graceful shutdown).
func (s *Service) Start() (<-chan error, error) {
	listener, err := (&net.ListenConfig{}).Listen( //nolint:exhaustruct_v5 // zero config is intentional
		context.Background(), "tcp", s.cfg.Addr)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "server.listen_failed", "listen on %s", s.cfg.Addr)
	}

	s.mu.Lock()
	s.ln = listener
	s.mu.Unlock()

	errCh := make(chan error, 1)

	go func() {
		serveErr := s.server.Serve(listener)
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
	errCh, err := s.Start() //nolint:contextcheck // Start is context-free by API design; Run owns the lifecycle ctx
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

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.cfg.ShutdownTimeout)
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

	// Drain hooks run while the server still serves traffic: readiness is
	// already down everywhere, so external readiness signals mounted on the
	// service (e.g. the health module's probe) flip in lockstep before the
	// drain wait gives load balancers time to observe the change.
	drainHooksErr := s.runDrainHooks(ctx)

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
		err = errorfamily.WrapInfrastructuref(err, "server.shutdown_failed", "shutdown failed")
	}

	// Hooks run after the server released its connections so telemetry
	// providers flush spans covering the final in-flight requests.
	hooksErr := s.runShutdownHooks(ctx)

	return errors.Join(drainHooksErr, err, hooksErr)
}

// runDrainHooks invokes each configured DrainHook in order with the shutdown
// context and joins the errors. Hooks run at most once per service: Shutdown
// is a no-op after the first call.
func (s *Service) runDrainHooks(ctx context.Context) error {
	var errs []error

	for _, hook := range s.cfg.DrainHooks {
		err := hook(ctx)
		if err != nil {
			errs = append(errs, errorfamily.WrapInfrastructuref(
				err, "server.drain_hook_failed", "drain hook failed",
			))
		}
	}

	return errors.Join(errs...)
}

// runShutdownHooks invokes each configured ShutdownHook in order with the
// shutdown context and joins the errors. Hooks run at most once per service:
// Shutdown is a no-op after the first call.
func (s *Service) runShutdownHooks(ctx context.Context) error {
	var errs []error

	for _, hook := range s.cfg.ShutdownHooks {
		err := hook(ctx)
		if err != nil {
			errs = append(errs, errorfamily.WrapInfrastructuref(
				err, "server.shutdown_hook_failed", "shutdown hook failed",
			))
		}
	}

	return errors.Join(errs...)
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
