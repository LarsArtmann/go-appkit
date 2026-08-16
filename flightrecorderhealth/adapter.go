package flightrecorderhealth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	fr "github.com/larsartmann/go-flightrecorder"
	"github.com/samber/do/v2"
)

// Checkable is a health-checkable wrapper around a flight recorder. It
// implements [do.HealthcheckerWithContext] so that when registered in a
// samber/do injector, the flight recorder's own operational state appears as
// a row in the health dashboard.
//
// A Checkable reports healthy when the underlying recorder is enabled (i.e.,
// the trace buffer is actively recording). A disabled or nil recorder reports
// unhealthy, signaling that trace capture is not functioning.
type Checkable struct {
	rec  *fr.Recorder
	name string
}

// CheckableOption configures a [Checkable].
type CheckableOption func(*Checkable)

// WithCheckableName sets the display name for this service in the health
// dashboard. Default: "flight-recorder".
func WithCheckableName(name string) CheckableOption {
	return func(c *Checkable) { c.name = name }
}

// NewCheckable wraps a flight recorder as a health-checkable service. The
// returned value implements [do.HealthcheckerWithContext] and should be
// registered in the samber/do injector.
//
// If rec is nil, HealthCheck returns an error immediately. This is a
// programming error — do not construct a Checkable without a recorder.
func NewCheckable(rec *fr.Recorder, opts ...CheckableOption) *Checkable {
	c := &Checkable{
		rec:  rec,
		name: "flight-recorder",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// HealthCheck reports whether the flight recorder is actively recording.
// Returns nil if enabled, an error otherwise.
func (c *Checkable) HealthCheck(_ context.Context) error {
	if c == nil || c.rec == nil {
		return fmt.Errorf("flightrecorder: recorder is not configured")
	}

	if !c.rec.Enabled() {
		return fmt.Errorf("flightrecorder: recorder is not enabled (trace buffer inactive)")
	}

	return nil
}

// Name returns the service name. Satisfies the optional name provider
// pattern used by samber/do for display purposes.
func (c *Checkable) Name() string {
	if c == nil {
		return ""
	}

	return c.name
}

// Trigger is a [health.HealthRecorder] that intercepts every health-check
// batch and triggers a flight recorder snapshot when any service fails. The
// snapshot captures the trace window around the failure for offline analysis
// with `go tool trace`.
//
// The trigger uses [fr.SnapshotIfAsync] so the capture is non-blocking — the
// health-check loop is never delayed by trace I/O.
type Trigger struct {
	rec          *fr.Recorder
	triggerFunc  fr.TriggerFunc
	logger       *slog.Logger
	cooldown     time.Duration
	lastCapture  time.Time
}

// TriggerOption configures a [Trigger].
type TriggerOption func(*Trigger)

// WithTriggerFunc sets the flight recorder trigger function that determines
// whether to capture a snapshot. Default: [fr.OnError] (captures when any
// health check returns an error).
//
// Use [fr.OnErrorOrLatency] to also capture on slow health checks, or
// [fr.OnAlways] to capture on every batch.
func WithTriggerFunc(trigger fr.TriggerFunc) TriggerOption {
	return func(t *Trigger) { t.triggerFunc = trigger }
}

// WithTriggerLogger sets the slog logger for snapshot capture events. When
// set, the trigger logs each capture with the failing service names and
// errors. Default: no logging.
func WithTriggerLogger(logger *slog.Logger) TriggerOption {
	return func(t *Trigger) { t.logger = logger }
}

// WithCooldown sets the minimum duration between captures. When non-zero,
// snapshot requests within the cooldown window are silently dropped. This
// prevents trace flooding when a service flaps repeatedly.
//
// Default: 0 (no cooldown — every failing batch triggers a capture, subject
// to the recorder's own once-semantics and sink configuration).
//
// For directory-sink recorders ([fr.WithSnapshotDir]), a cooldown of 30s–60s
// is recommended to avoid excessive trace files. For writer-sink recorders,
// the once-latch already prevents repeated captures; cooldown is unnecessary.
func WithCooldown(d time.Duration) TriggerOption {
	return func(t *Trigger) { t.cooldown = d }
}

// NewTrigger creates a [health.HealthRecorder] that captures a flight recorder
// snapshot when health checks fail. The snapshot is captured asynchronously
// via [fr.SnapshotIfAsync], so the health-check loop is never blocked.
//
// If rec is nil, RecordHealthCheckWithContext is a pass-through that delegates
// to the injector directly. This allows construction before the recorder is
// available (e.g., config-gated disabling).
func NewTrigger(rec *fr.Recorder, opts ...TriggerOption) *Trigger {
	t := &Trigger{
		rec:         rec,
		triggerFunc: fr.OnError(),
		cooldown:    0,
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// RecordHealthCheckWithContext satisfies [health.HealthRecorder]. It runs the
// health-check batch via the injector, then evaluates the trigger function
// to decide whether to capture a trace snapshot. The snapshot is captured
// asynchronously via [fr.SnapshotIfAsync], so the health-check loop is never
// blocked.
//
// The [fr.TriggerContext].Err is set to the first failing service's error
// (nil when all services pass). This lets [fr.OnError] fire only on failures
// while [fr.OnAlways] fires on every batch.
func (t *Trigger) RecordHealthCheckWithContext(
	ctx context.Context,
	injector do.Injector,
) map[string]error {
	if t == nil || t.rec == nil {
		return injector.HealthCheckWithContext(ctx)
	}

	results := injector.HealthCheckWithContext(ctx)

	tc := fr.TriggerContext{
		Kind: "health.check",
		Type: "batch",
		Err:  firstError(results),
	}

	// Cooldown check — compare against the last capture time.
	if t.cooldown > 0 && !t.lastCapture.IsZero() {
		if time.Since(t.lastCapture) < t.cooldown {
			return results
		}
	}

	captured := t.rec.SnapshotIfAsync(context.WithoutCancel(ctx), tc, t.triggerFunc)

	if captured {
		t.lastCapture = time.Now()

		if t.logger != nil {
			t.logger.InfoContext(ctx,
				"flightrecorder: trace snapshot triggered by health check",
				"failed_services", failingServiceNames(results),
			)
		}
	}

	return results
}

// failingServiceNames returns the names of all services with non-nil errors.
func failingServiceNames(results map[string]error) []string {
	names := make([]string, 0, len(results))

	for name, err := range results {
		if err != nil {
			names = append(names, name)
		}
	}

	return names
}

// firstError returns the first non-nil error from the results map, or nil if
// all services passed. This populates [fr.TriggerContext].Err so that
// [fr.OnError] and [fr.OnErrorOrLatency] triggers fire on health-check
// failures.
func firstError(results map[string]error) error {
	for _, err := range results {
		if err != nil {
			return err
		}
	}

	return nil
}

// Register is a convenience function that creates a [Checkable] and registers
// it in the samber/do injector as a named service. The health Probe will
// discover it automatically.
//
//	rec, _ := flightrecorder.New(flightrecorder.WithSnapshotDir("/var/traces"))
//	flightrecorderhealth.Register(injector, rec, "flight-recorder")
//
// If you need the HealthRecorder trigger as well, use [NewTrigger] separately
// and pass it to [health.New] via [health.WithHealthRecorder].
func Register(injector do.Injector, rec *fr.Recorder, name string, opts ...CheckableOption) *Checkable {
	if name == "" {
		name = "flight-recorder"
	}

	c := NewCheckable(rec, append(opts, WithCheckableName(name))...)

	do.ProvideNamed(injector, name, func(_ do.Injector) (*Checkable, error) {
		return c, nil
	})

	// Eagerly invoke to instantiate the service in the container.
	_, _ = do.InvokeNamed[*Checkable](injector, name)

	return c
}
