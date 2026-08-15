// Package cqrs provides CQRS/ES integration for go-appkit services.
// It wraps go-cqrs-lite/stack/sqlite and projectionhost into a lifecycle-managed
// EventService that integrates with appkit.Service for graceful shutdown.
package cqrs

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/larsartmann/go-cqrs-lite/flightrecorder/v4"
	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	"github.com/larsartmann/go-cqrs-lite/stack/sqlite/v4"
	stack "github.com/larsartmann/go-cqrs-lite/stack/v4"
	errorfamily "github.com/larsartmann/go-error-family"
)

// EventConfig configures the CQRS event service.
type EventConfig struct {
	// SQLitePath is the path to the SQLite database file.
	// Required — must not be empty.
	SQLitePath string

	// StackOptions are passed through to stack/sqlite.New.
	// Use these to customize WAL, foreign keys, optimizations, etc.
	StackOptions []sqlite.Option

	// Logger receives projection host lifecycle events: worker crashes,
	// restarts, dead-letter captures, and shutdowns. Wire the same logger
	// you hand to appkit.Service so projection trouble lands in one place.
	// Default: slog.Default().
	Logger *slog.Logger

	// DLQ enables poison-event capture for projections. When nil (default),
	// a repeatedly failing event eventually exhausts the worker's restart
	// budget and the projection stalls in WorkerFailed. When set, events that
	// fail more than DLQ.Threshold times are quarantined in a dead-letter
	// store and the checkpoint advances — one poison event cannot stall a
	// projection. See DLQConfig for store and threshold defaults.
	DLQ *DLQConfig

	// FlightRecorder captures a runtime/trace snapshot when a projection
	// worker exhausts its restart budget and transitions to WorkerFailed —
	// terminal failures are rare and high-signal, so every one is captured
	// (OnAlways). The recorder must be started separately, typically at
	// application startup; snapshots go to the recorder's configured writer.
	// Go allows only ONE active flight recorder per process: coordinate with
	// the HTTP-level github.com/larsartmann/go-flightrecorder used by the
	// appkit/flightrecorder middleware module if you are tempted to run both.
	FlightRecorder *flightrecorder.Recorder

	// Metrics observes projection host lifecycle events: processed and
	// errored events, dead-letter captures, worker restarts and terminal
	// failures, and checkpoint advance with lag. Implementations must be
	// safe for concurrent use and must not block (the host records
	// fire-and-forget from every worker goroutine). The interface is
	// backend-agnostic — forward the calls to Prometheus, OTel, or any
	// stats sink. Nil (default) disables metrics.
	Metrics projectionhost.MetricsRecorder

	// HostOptions are passed through to projectionhost.New for advanced
	// tuning (WithMaxRestarts, WithBackoff, WithBatchSize,
	// WithShutdownTimeout, ...). Options derived from Logger, Metrics,
	// and FlightRecorder are appended after these, so derived wiring wins
	// conflicts. (DLQ wiring is derived in NewEventService.)
	HostOptions []projectionhost.HostOption
}

// DLQConfig configures the projection dead-letter queue.
type DLQConfig struct {
	// Threshold is the number of handler failures before an event is
	// quarantined. Values <= 0 keep the projectionhost default (3).
	Threshold int

	// Store persists dead-letter entries. When nil (default), a SQLite-backed
	// store is created in the event store's own database (table
	// projection_dead_letters), so entries survive restarts. Provide
	// projectionhost.NewMemoryDeadLetterStore for ephemeral tests.
	Store projectionhost.DeadLetterStore
}

// EventService manages a CQRS/ES event store backed by SQLite.
// It wraps stack/sqlite.Bundle and projectionhost.Host with lifecycle
// management that integrates with appkit.Service.
type EventService struct {
	bundle *stack.Bundle
	host   *projectionhost.Host
	dlq    projectionhost.DeadLetterStore
	mu     sync.Mutex
	closed bool
}

// NewEventService creates an EventService from the given config.
// The SQLite database is opened, schema is auto-migrated, and the
// projection host is initialized (but not started).
func NewEventService(cfg EventConfig) (*EventService, error) {
	if cfg.SQLitePath == "" {
		return nil, errorfamily.NewRejection("cqrs.path_required", "SQLitePath is required")
	}

	bundle, err := sqlite.New(cfg.SQLitePath, cfg.StackOptions...)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(
			err,
			"cqrs.open_failed",
			"failed to open event store at %s",
			cfg.SQLitePath,
		)
	}

	hostOpts := cfg.hostOptions()

	dlqStore, dlqErr := resolveDLQ(cfg.DLQ, bundle)
	if dlqErr != nil {
		_ = bundle.GracefulClose(context.Background())

		return nil, dlqErr
	}

	if dlqStore != nil {
		hostOpts = append(hostOpts,
			projectionhost.WithDeadLetterStore(dlqStore, cfg.DLQ.Threshold))
	}

	host, err := projectionhost.New(
		bundle.SeekableJournal,
		bundle.CheckpointStore,
		hostOpts...,
	)
	if err != nil {
		_ = bundle.GracefulClose(context.Background())

		return nil, errorfamily.WrapInfrastructuref(
			err,
			"cqrs.projection_host_failed",
			"failed to create projection host",
		)
	}

	return &EventService{
		bundle: bundle,
		host:   host,
		dlq:    dlqStore,
	}, nil
}

// resolveDLQ determines the dead-letter store for a config. A nil cfg or a
// cfg with an explicit Store needs no bundle access; the default store is
// provisioned in the bundle's own database.
func resolveDLQ(
	cfg *DLQConfig,
	bundle *stack.Bundle,
) (projectionhost.DeadLetterStore, error) {
	if cfg == nil {
		return nil, nil
	}

	if cfg.Store != nil {
		return cfg.Store, nil
	}

	sqlDB, err := asSQLDB(bundle.Database())
	if err != nil {
		return nil, errorfamily.WrapInfrastructure(
			err,
			"cqrs.dlq_db_unavailable",
			"cannot provision SQLite dead-letter store",
		)
	}

	store, err := projectionhost.NewSQLiteDeadLetterStore(context.Background(), sqlDB)
	if err != nil {
		return nil, errorfamily.WrapInfrastructuref(
			err,
			"cqrs.dlq_provision_failed",
			"failed to create dead-letter store",
		)
	}

	return store, nil
}

// hostOptions maps EventConfig onto projectionhost options.
// Nil-valued config fields are skipped so projectionhost defaults apply.
// Consumer-supplied HostOptions come first; derived wiring wins conflicts.
func (cfg EventConfig) hostOptions() []projectionhost.HostOption {
	opts := append([]projectionhost.HostOption{}, cfg.HostOptions...)

	if cfg.Logger != nil {
		opts = append(opts, projectionhost.WithLogger(cfg.Logger))
	}

	if cfg.FlightRecorder != nil {
		opts = append(opts, projectionhost.WithFlightRecorder(cfg.FlightRecorder, nil))
	}

	if cfg.Metrics != nil {
		opts = append(opts, projectionhost.WithMetrics(cfg.Metrics))
	}

	return opts
}

// Bundle returns the underlying stack.Bundle.
// Use this to access EventSink, EventSource, CommandSink, Publisher, etc.
func (es *EventService) Bundle() *stack.Bundle {
	return es.bundle
}

// Host returns the underlying projectionhost.Host.
// Use this to register projections before starting the service.
func (es *EventService) Host() *projectionhost.Host {
	return es.host
}

// DeadLetterStore returns the configured dead-letter store, or nil when the
// DLQ is disabled. The SQLite default additionally implements
// projectionhost.DeadLetterStoreAdmin (Count, ListPaged, PurgeBefore) via
// type assertion.
func (es *EventService) DeadLetterStore() projectionhost.DeadLetterStore {
	return es.dlq
}

// ReplayDeadLetters re-feeds dead-letter entries to their registered
// projections. It is a pure retry: fix the handler bug first, then call this;
// successful entries are reported in ReplayResult.Replayed and must be
// removed from the store by the caller (Store.Delete or Purge). The host need
// not be running. Returns an error when the DLQ is disabled.
func (es *EventService) ReplayDeadLetters(
	ctx context.Context,
	projectionName string,
) (projectionhost.ReplayResult, error) {
	if es.dlq == nil {
		return projectionhost.ReplayResult{},
			errorfamily.NewRejection("cqrs.dlq_disabled", "DLQ is not configured")
	}

	return es.host.ReplayDeadLetters(ctx, projectionName)
}

// ResetProjection rewinds a projection's checkpoint to the beginning (or to a
// specific event with reset options) so it reprocesses its stream. Pass
// projectionhost.WithPurgeDeadLetters() to also clear its dead-letter
// entries. Use after fixing a handler bug to rebuild a projection's state.
func (es *EventService) ResetProjection(
	ctx context.Context,
	name string,
	opts ...projectionhost.ResetOption,
) error {
	return es.host.Reset(ctx, name, opts...)
}

// DB extracts the *sql.DB from the bundle's Database() method.
// Returns a Rejection error if the database is not backed by *sql.DB.
func (es *EventService) DB() (*sql.DB, error) {
	return asSQLDB(es.bundle.Database())
}

// asSQLDB type-asserts a stack bundle database handle to *sql.DB.
func asSQLDB(bundleDB any) (*sql.DB, error) {
	sqlDB, ok := bundleDB.(*sql.DB)
	if !ok {
		return nil, errorfamily.Newf(
			errorfamily.Rejection,
			"cqrs.db_not_sql",
			"database is not *sql.DB (got %T)",
			bundleDB,
		)
	}

	return sqlDB, nil
}

// ReadyCheck reports whether all registered projections are serving: every
// worker must be live (caught up and processing) or stopped (fully drained,
// normal for batch-style hosts). A worker that is idle before
// StartProjections, still catching up, backing off after a crash, draining,
// or terminally failed makes the service NOT ready. Wire it into
// appkit.ServiceConfig.ReadyCheck so /health/ready serves 503 until
// projections are caught up and flips back if one dies:
//
//	cfg.ReadyCheck = eventSvc.ReadyCheck
//
// With no registered projections it reports true.
func (es *EventService) ReadyCheck() bool {
	for _, state := range es.host.Status() {
		switch state.Status {
		case projectionhost.WorkerLive, projectionhost.WorkerStopped:
			continue
		case projectionhost.WorkerIdle,
			projectionhost.WorkerRunning,
			projectionhost.WorkerBackoff,
			projectionhost.WorkerDraining,
			projectionhost.WorkerFailed:
			return false
		}
	}

	return true
}

// LagPerProjection reports how far behind real-time each projection is, keyed
// by projection name. Useful for dashboards and alerting; the same data
// drives staleness decisions in production.
func (es *EventService) LagPerProjection() map[string]time.Duration {
	return es.host.LagPerProjection()
}

// StartProjections starts the projection host workers.
// Must be called after all projections are registered and before the service begins serving.
func (es *EventService) StartProjections(ctx context.Context) error {
	return es.host.Start(ctx)
}

// Shutdown gracefully stops projections and closes the event store.
// Safe to call multiple times (idempotent via mutex guard).
func (es *EventService) Shutdown(ctx context.Context) error {
	es.mu.Lock()

	if es.closed {
		es.mu.Unlock()

		return nil
	}

	es.closed = true
	es.mu.Unlock()

	return errors.Join(es.host.Stop(), es.bundle.GracefulClose(ctx))
}
