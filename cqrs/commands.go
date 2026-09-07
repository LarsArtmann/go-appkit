// Command/query facade on EventService: typed registration and dispatch
// passthroughs to the underlying system.System, the default middleware
// builder, and the in-flight command drain used by Shutdown.
package cqrs

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/middleware/v4"
	cqrsotel "github.com/larsartmann/go-cqrs-lite/otel/v4"
	"github.com/larsartmann/go-cqrs-lite/query/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
)

// DefaultCommandMiddleware composes a sane command chain: recovery,
// optional OTel tracing (omit when tracer is nil), and logging. Retry,
// idempotency, and circuit breaking are deliberately NOT defaults — they
// change command semantics and belong to the consumer (append them via
// EventConfig.CommandMiddleware).
func DefaultCommandMiddleware(logger *slog.Logger, tracer cqrsotel.Tracer) []command.Middleware {
	chain := []command.Middleware{middleware.CommandRecovery()}

	if tracer != nil {
		chain = append(chain, middleware.CommandTracing(tracer))
	}

	if logger != nil {
		chain = append(chain, middleware.CommandLogging(logger))
	}

	return chain
}

// RegisterDecider registers a decider for a stream type (typed passthrough
// to system.RegisterDecider). Must be called before the matching
// RegisterCommand.
func RegisterDecider[State any](
	es *EventService,
	streamType string,
	d decider.Decider[State],
	opts ...system.RegisterDeciderOption,
) error {
	return system.RegisterDecider(es.sys, streamType, d, opts...) //nolint:wrapcheck // delegation
}

// RegisterCommand registers a typed command handler that returns an
// system.Op (typed passthrough to system.RegisterCommand).
func RegisterCommand[Cmd command.Command, State any](
	es *EventService,
	name command.Type,
	handler func(ctx context.Context, cmd Cmd) system.Op[State],
) error {
	return system.RegisterCommand[Cmd, State](es.sys, name, handler) //nolint:wrapcheck // delegation
}

// RegisterQuery registers a typed query handler (typed passthrough to
// system.RegisterQuery).
func RegisterQuery[Q any, R any](
	es *EventService,
	name string,
	handler func(ctx context.Context, q Q) (R, error),
) error {
	return system.RegisterQuery[Q, R](es.sys, name, handler) //nolint:wrapcheck // delegation
}

// Dispatch sends a command through the service's command dispatcher
// (domain middleware included). The handler must have been registered via
// RegisterCommand.
func (es *EventService) Dispatch(ctx context.Context, cmd command.Command) error {
	return es.sys.CommandDispatcher().Dispatch(ctx, cmd) //nolint:wrapcheck // delegation
}

// DispatchQuery dispatches a typed query and returns the result (typed
// passthrough to system.DispatchQuery).
func DispatchQuery[Q query.Query, R any](ctx context.Context, es *EventService, q Q) (R, error) {
	return system.DispatchQuery[Q, R](ctx, es.sys, q) //nolint:wrapcheck // delegation
}

// DispatchQueryChecked guards a typed query with the read-your-writes
// staleness check: when the maximum projection lag exceeds maxStaleness the
// query is NOT answered — the staleness error (Transient family, ideal for
// 503s) is returned instead. A maxStaleness <= 0 disables the check.
func DispatchQueryChecked[Q query.Query, R any](
	ctx context.Context,
	es *EventService,
	maxStaleness time.Duration,
	q Q,
) (R, error) {
	var zero R

	if err := es.CheckStaleness(maxStaleness); err != nil {
		return zero, err
	}

	return DispatchQuery[Q, R](ctx, es, q)
}

// CommandDispatcher exposes the raw command dispatcher for advanced wiring
// (publish middleware, per-name inspection).
func (es *EventService) CommandDispatcher() *command.Dispatcher {
	return es.sys.CommandDispatcher()
}

// QueryDispatcher exposes the raw query dispatcher for advanced wiring.
func (es *EventService) QueryDispatcher() *query.Dispatcher {
	return es.sys.QueryDispatcher()
}

// inFlightTracker counts commands executing through the outermost
// middleware slot so Shutdown can wait for them before closing engines.
type inFlightTracker struct {
	mu      sync.Mutex
	pending int
}

// newInFlightTracker creates a tracker.
func newInFlightTracker() *inFlightTracker {
	return &inFlightTracker{}
}

// commandMiddleware returns the outermost tracking middleware.
func (t *inFlightTracker) commandMiddleware() command.Middleware {
	return func(next command.Handler) command.Handler {
		return func(ctx context.Context, cmd command.Command) error {
			t.enter()
			defer t.exit()

			return next(ctx, cmd)
		}
	}
}

// enter records an in-flight command.
func (t *inFlightTracker) enter() {
	t.mu.Lock()
	t.pending++
	t.mu.Unlock()
}

// exit records a completed command.
func (t *inFlightTracker) exit() {
	t.mu.Lock()
	t.pending--
	t.mu.Unlock()
}

// pendingCount snapshots the in-flight count.
func (t *inFlightTracker) pendingCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.pending
}

// drain blocks until all in-flight commands complete or the context
// expires. It never aborts running commands — that is the engines' job via
// context cancellation in GracefulClose.
func (t *inFlightTracker) drain(ctx context.Context) error {
	ticker := time.NewTicker(2 * time.Millisecond)
	defer ticker.Stop()

	for t.pendingCount() > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}

	return nil
}
