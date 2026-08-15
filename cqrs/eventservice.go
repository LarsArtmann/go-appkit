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
}

// EventService manages a CQRS/ES event store backed by SQLite.
// It wraps stack/sqlite.Bundle and projectionhost.Host with lifecycle
// management that integrates with appkit.Service.
type EventService struct {
	bundle *stack.Bundle
	host   *projectionhost.Host
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

	host, err := projectionhost.New(
		bundle.SeekableJournal,
		bundle.CheckpointStore,
		cfg.hostOptions()...,
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
	}, nil
}

// hostOptions maps EventConfig onto projectionhost options.
// Nil-valued config fields are skipped so projectionhost defaults apply.
func (cfg EventConfig) hostOptions() []projectionhost.HostOption {
	var opts []projectionhost.HostOption

	if cfg.Logger != nil {
		opts = append(opts, projectionhost.WithLogger(cfg.Logger))
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

// DB extracts the *sql.DB from the bundle's Database() method.
// Returns a Rejection error if the database is not backed by *sql.DB.
func (es *EventService) DB() (*sql.DB, error) {
	return asSQLDB(es.bundle.Database())
}

// asSQLDB type-asserts a stack bundle database handle to *sql.DB.
func asSQLDB(db any) (*sql.DB, error) {
	sqlDB, ok := db.(*sql.DB)
	if !ok {
		return nil, errorfamily.Newf(
			errorfamily.Rejection,
			"cqrs.db_not_sql",
			"database is not *sql.DB (got %T)",
			db,
		)
	}

	return sqlDB, nil
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
