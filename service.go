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

// activeServe is the current serve's delegated http.Server ownership: appkit
// binds and classifies, httputil owns the server loop (timeouts, graceful
// drain, the future TLS seam). A fresh instance is built for every Start —
// httputil's single-serve ownership consumes the listener when a serve ends
// (server.listener_consumed), so a handle is never reused across serves.
type activeServe struct {
	srv *httputil.Server

	// graceful is closed by Shutdown once the delegated Shutdown call has
	// returned: the signal resolves the nil-on-graceful branch of the serve
	// errCh contract, which httputil's error channel cannot express (it only
	// delivers failures).
	graceful chan struct{}
}

// Service is a production-ready HTTP service with middleware, health checks,
// structured logging, and graceful shutdown.
type Service struct {
	cfg ServiceConfig

	Logger *slog.Logger
	Mux    *http.ServeMux

	handler    http.Handler
	ln         net.Listener
	serve      *activeServe
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

	svc := &Service{ //nolint:exhaustruct_v5 // handler, ln, serve, readyProbe, mu are deliberate zero values
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

	var collector *metricsCollector

	if cfg.Metrics != nil {
		cfg.Metrics.applyMetricsDefaults()

		if err := cfg.Metrics.validate(); err != nil {
			return nil, err
		}

		collector = newMetricsCollector(cfg.Version)
		mux.Handle("GET "+cfg.Metrics.Path, metricsAuth(cfg.Metrics, collector.handler()))
	}

	if cfg.Version != "" {
		mux.HandleFunc("GET /version", versionHandler(cfg.Version))
	}

	mws := buildMiddleware(logger, cfg)
	if collector != nil {
		// The collector wraps the mux directly (inside the default stack's
		// outermost layer) so it observes route patterns stamped by the mux
		// while still counting requests that the default middlewares admit.
		mws = append(mws, collector.middleware)
	}

	wrapped := httputil.Chain(mux, mws...)
	svc.handler = wrapped

	return svc, nil
}

// serverConfig maps the service config onto the delegated server's config:
// every timeout field passes through unchanged except the NoTimeout sentinel,
// which becomes 0 (the stdlib's "no deadline"), and ShutdownTimeout, which
// carries the SAME config field that Shutdown uses for its context deadline
// so the effective drain deadline has exactly one source of truth.
func (s *Service) serverConfig(addr string) httputil.ServerConfig {
	return httputil.ServerConfig{
		Addr:              addr,
		ReadTimeout:       serverTimeout(s.cfg.ReadTimeout),
		ReadHeaderTimeout: s.cfg.ReadHeaderTimeout,
		WriteTimeout:      serverTimeout(s.cfg.WriteTimeout),
		IdleTimeout:       s.cfg.IdleTimeout,
		ShutdownTimeout:   s.cfg.ShutdownTimeout,
	}
}

// Start creates the listener, runs the StartHooks (post-listen, pre-serve),
// and begins serving in a goroutine. Returns a synchronous error if the
// listener cannot bind or a StartHook fails — on hook failure the listener
// is closed again and the service never serves. The returned channel
// receives any serve error (nil on graceful shutdown).
func (s *Service) Start() (<-chan error, error) {
	listener, err := (&net.ListenConfig{}).Listen( //nolint:exhaustruct_v5 // zero config is intentional
		context.Background(), "tcp", s.cfg.Addr)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(err, "server.listen_failed", "listen on %s", s.cfg.Addr)
	}

	// The delegated server is constructed before the hooks: NewServerListener
	// is pure (it validates and wraps, never binds or serves), so a config
	// the delegated server rejects surfaces synchronously here, before any
	// observable lifecycle state exists. The listener stays untouched on
	// error and is closed again below.
	delegated, err := httputil.NewServerListener(listener, s.serverConfig(listener.Addr().String()), s.handler)
	if err != nil {
		_ = listener.Close()

		return nil, errorfamily.Newf(
			errorfamily.Rejection,
			"server.config_invalid",
			"server configuration rejected: %v",
			err,
		)
	}

	s.mu.Lock()
	s.ln = listener
	s.serve = &activeServe{srv: delegated, graceful: make(chan struct{})}
	s.mu.Unlock()

	// Start hooks run after the listener is bound (Addr and Running
	// already observe it) but before the serve loop accepts traffic — the
	// startup mirror of the drain window. Unlike the shutdown hooks, any
	// failure fails the whole start: a half-announced service must never
	// serve. The phase line mirrors the shutdown phase-log contract.
	phaseStart := time.Now()

	hooksErr := s.runStartHooks(context.Background())
	if hooksErr != nil {
		s.mu.Lock()
		s.ln = nil
		s.serve = nil
		s.mu.Unlock()

		_ = listener.Close()

		return nil, hooksErr
	}

	s.logStartupPhase("start_hooks", phaseStart)

	errCh := make(chan error, 1)
	rawCh := delegated.Start()
	graceful := s.serve.graceful

	go func() {
		// httputil's channel delivers ONLY serve failures (a graceful close
		// sends nothing), so the graceful branch is driven by Shutdown closing
		// the graceful channel. The non-blocking peek first gives a failure
		// that already arrived priority over a graceful signal closing in the
		// same instant; after the peek, whichever fires first wins, and the
		// buffered send guarantees exactly one value on errCh.
		select {
		case serveErr := <-rawCh:
			errCh <- errorfamily.WrapInfrastructuref(serveErr, "server.serve_failed", "serve failed")

			return
		default:
		}

		select {
		case serveErr := <-rawCh:
			errCh <- errorfamily.WrapInfrastructuref(serveErr, "server.serve_failed", "serve failed")
		case <-graceful:
			errCh <- nil
		}
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
// 2. Run DrainHooks while traffic is still served
// 3. Wait DrainDelay for load balancer propagation
// 4. Stop accepting new connections and finish in-flight requests
// 5. Run ShutdownHooks after connections are released.
//
// Every phase emits one INFO line ("shutdown phase complete" carrying the
// phase name and its duration), so a deploy can be diagnosed from logs
// alone; a final "graceful shutdown complete" line reports the total.
func (s *Service) Shutdown(ctx context.Context) error {
	totalStart := time.Now()

	s.mu.Lock()
	ln := s.ln
	serve := s.serve
	s.ln = nil
	s.serve = nil
	s.mu.Unlock()

	if ln == nil {
		return nil
	}

	phaseStart := time.Now()

	s.readyProbe.Store(false)
	s.logPhase("ready_flip", phaseStart)

	// Drain hooks run while the server still serves traffic: readiness is
	// already down everywhere, so external readiness signals mounted on the
	// service (e.g. the health module's probe) flip in lockstep before the
	// drain wait gives load balancers time to observe the change.
	phaseStart = time.Now()
	drainHooksErr := s.runDrainHooks(ctx)
	s.logPhase("drain_hooks", phaseStart)

	if s.cfg.DrainDelay > 0 {
		s.Logger.Info("draining traffic", "delay", s.cfg.DrainDelay)

		phaseStart = time.Now()

		timer := time.NewTimer(s.cfg.DrainDelay)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-ctx.Done():
		}

		s.logPhase("drain_wait", phaseStart)
	} else {
		s.Logger.Info("shutdown phase skipped", "phase", "drain_wait")
	}

	phaseStart = time.Now()

	err := serve.srv.Shutdown(ctx)
	if err != nil {
		err = errorfamily.WrapInfrastructuref(err, "server.shutdown_failed", "shutdown failed")
	}

	// Closed regardless of the shutdown result: even a deadline-expired
	// delegated Shutdown has already closed the listener, so the serve loop
	// ends with http.ErrServerClosed (which httputil's channel never reports)
	// and the nil-on-graceful branch must resolve exactly as on a clean stop.
	close(serve.graceful)

	s.logPhase("listener_close", phaseStart)

	// Hooks run after the server released its connections so telemetry
	// providers flush spans covering the final in-flight requests.
	phaseStart = time.Now()
	hooksErr := s.runShutdownHooks(ctx)
	s.logPhase("shutdown_hooks", phaseStart)

	joined := errors.Join(drainHooksErr, err, hooksErr)

	result := "ok"
	if joined != nil {
		result = "error"
	}

	s.Logger.Info("graceful shutdown complete", "total", time.Since(totalStart), "result", result)

	return joined
}

// logPhase emits the per-phase shutdown line: a stable message carrying the
// phase name and its duration.
func (s *Service) logPhase(phase string, start time.Time) {
	s.Logger.Info("shutdown phase complete", "phase", phase, "duration", time.Since(start))
}

// logStartupPhase emits the startup phase line, mirroring logPhase so a
// boot can be diagnosed from logs alone. Emitted only when the phase
// succeeded — a failed startup surfaces through Start's classified error
// instead, and a service that never serves must not log completion lines.
func (s *Service) logStartupPhase(phase string, start time.Time) {
	s.Logger.Info("startup phase complete", "phase", phase, "duration", time.Since(start))
}

// runStartHooks invokes each configured StartHook in order and joins the
// errors; Start aborts on any failure, closing the listener again.
func (s *Service) runStartHooks(ctx context.Context) error {
	return s.runHooks(ctx, s.cfg.StartHooks, "server.start_hook_failed", "start hook failed")
}

// runDrainHooks invokes each configured DrainHook in order with the shutdown
// context and joins the errors. Hooks run at most once per service: Shutdown
// is a no-op after the first call.
func (s *Service) runDrainHooks(ctx context.Context) error {
	return s.runHooks(ctx, s.cfg.DrainHooks, "server.drain_hook_failed", "drain hook failed")
}

// runShutdownHooks invokes each configured ShutdownHook in order with the
// shutdown context and joins the errors. Hooks run at most once per service:
// Shutdown is a no-op after the first call.
func (s *Service) runShutdownHooks(ctx context.Context) error {
	return s.runHooks(ctx, s.cfg.ShutdownHooks, "server.shutdown_hook_failed", "shutdown hook failed")
}

// runHooks invokes each hook in order, letting every hook run even when an
// earlier one fails, and joins each failure as an Infrastructure error under
// the phase-specific code consumers match on. A hook that already returns an
// errorfamily error keeps its own classification (WrapOnce — no re-wrap
// nesting); foreign errors gain the phase code, and every failure carries
// its hook position in the error context.
func (s *Service) runHooks(ctx context.Context, hooks []Hook, code, message string) error {
	var errs []error

	for i, hook := range hooks {
		err := hook(ctx)
		if err != nil {
			errs = append(errs, errorfamily.WrapOnce(err, errorfamily.Infrastructure, code, message).
				WithContextf("hook_index", "%d", i))
		}
	}

	return errors.Join(errs...)
}

// Close calls Shutdown with context.Background(). Idempotent.
func (s *Service) Close() error {
	return s.Shutdown(context.Background())
}

// Addr returns the bound listener address. Returns nil before Start is
// called AND from the moment Shutdown begins: Shutdown reaps the listener
// before running the drain hooks, so drain hooks (and everything after
// them) observe nil — capture the address before calling Shutdown.
func (s *Service) Addr() net.Addr {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.ln == nil {
		return nil
	}

	return s.ln.Addr()
}

// Running reports whether the service has a bound listener. It flips to
// false the moment Shutdown begins — before the drain hooks run — not
// when the last in-flight request finishes.
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
