// Package cqrs provides CQRS/ES integration for go-appkit services.
// It wraps go-cqrs-lite/system (the strategic composition layer) and
// projectionhost into a lifecycle-managed EventService that integrates with
// appkit.Service for graceful shutdown.
//
// v0.5.0 BREAKING: the engine room moved from the deprecated stack/sqlite
// preset (removed at go-cqrs-lite v5, ADR-0123) to system.New with a
// DomainConfig/DeploymentConfig split. Operators can swap engines at
// deployment time (Driver/DSN/Pragmas, or a koanf YAML config file);
// developers register commands, queries, and projections on the service.
//
// cqrs-lint:ignore(E014) async-by-design wrapper: read-your-writes is a read-time guard (CheckStaleness/CheckProjectionStaleness, see README), not a post-command drain
package cqrs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	_ "github.com/larsartmann/go-cqrs-lite/metaengine/sqliteengine/v4" // registers the "sqlite" driver (blank-import contract)
	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	"github.com/larsartmann/go-cqrs-lite/query/v4"
	"github.com/larsartmann/go-cqrs-lite/storage/v4/eventstore"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	errorfamily "github.com/larsartmann/go-error-family"
	fr "github.com/larsartmann/go-flightrecorder"
)

// Engine names and driver identifiers used by the default deployment.
const (
	defaultEngineName   = "primary"
	defaultSQLiteDriver = "sqlite"
	memoryDriver        = "memory"
	auxCloserName       = "appkit-cqrs-aux-db"
)

// EventConfig configures the CQRS event service.
//
// Storage resolves in this precedence: Deployment (a fully pre-loaded
// operator config) > ConfigPath (koanf YAML + CQRS_ env overrides) >
// Driver/DSN/Pragmas (defaults: sqlite driver, file-backed).
type EventConfig struct {
	// DSN is the database path or connection string. Required unless
	// ConfigPath or Deployment is set, or Driver is "memory" for an
	// in-process store (tests).
	DSN string

	// Driver is the metaengine driver name: "sqlite" (default), "memory",
	// "pebble", "postgres", ... Drivers beyond "sqlite" and "memory" must be
	// blank-imported by the consumer so they self-register — a missing
	// blank import fails construction with "unknown driver".
	Driver string

	// Pragmas are SQLite pragmas passed to the sqlite driver (e.g.
	// "journal_mode=WAL"). Ignored by other drivers.
	Pragmas []string

	// SQLitePath is the deprecated v0.4.0 alias for DSN with the sqlite
	// driver. It will be removed at v0.6.0.
	//
	// Deprecated: use DSN.
	SQLitePath string

	// ConfigPath loads the DeploymentConfig from a YAML file (system.LoadConfig:
	// koanf tags + CQRS_ env overrides, e.g. CQRS_ENGINES__PRIMARY__DRIVER).
	// When set, DSN/Driver/Pragmas/SQLitePath are ignored. Optional.
	ConfigPath string

	// Deployment is a fully pre-loaded operator config. When set it wins over
	// every other storage field. The config MUST declare a RoleProjections
	// instance so the projection host exists. Optional.
	Deployment *system.DeploymentConfig

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
	// worker exhausts its restart budget and transitions to WorkerFailed.
	// The type is github.com/larsartmann/go-flightrecorder — the same
	// recorder the appkit/flightrecorder middleware module uses, so ONE
	// shared recorder instance can serve both HTTP-level and
	// projection-level captures. Go still allows only ONE active flight
	// recorder per process; start it once, typically at application
	// startup — snapshots go to the recorder's configured writer.
	// Default trigger: capture on every terminal failure (rare and
	// high-signal); customize via FlightRecorderTrigger.
	FlightRecorder *fr.Recorder

	// FlightRecorderTrigger decides whether a terminal projection-worker
	// failure captures a trace snapshot. It receives an fr.TriggerContext
	// (Kind "projection", Type = projection name, Err = terminal error)
	// and is evaluated synchronously on the worker goroutine — keep it
	// fast. Nil (default) captures every terminal failure (fr.OnAlways).
	// Only consulted when FlightRecorder is set.
	FlightRecorderTrigger fr.TriggerFunc

	// Metrics observes projection host lifecycle events: processed and
	// errored events, dead-letter captures, worker restarts and terminal
	// failures, and checkpoint advance with lag. Implementations must be
	// safe for concurrent use and must not block (the host records
	// fire-and-forget from every worker goroutine). The interface is
	// backend-agnostic — forward the calls to Prometheus, OTel, or any
	// stats sink. Nil (default) disables metrics.
	Metrics projectionhost.MetricsRecorder

	// CheckpointStore overrides the projection checkpoint store. When nil
	// (default) a persistent SQL checkpoint store is created on the config's
	// own SQLite database (driver "sqlite" with a DSN); other drivers get
	// in-memory checkpoints (full replays after restart). Use this to force
	// a custom store for any driver.
	CheckpointStore event.CheckpointStore

	// CommandMiddleware wraps every command dispatched through the service
	// (see RegisterCommand). The in-flight drain tracker is installed
	// outermost automatically. Nil (default) installs no middleware —
	// compose a chain via DefaultCommandMiddleware and append your own.
	CommandMiddleware []command.Middleware

	// QueryMiddleware wraps every query dispatched through the service
	// (see RegisterQuery). Nil (default) installs no middleware.
	QueryMiddleware []query.Middleware

	// HostOptions are passed through to the projection host for advanced
	// tuning (WithMaxRestarts, WithBackoff, WithBatchSize,
	// WithShutdownTimeout, ...). Options derived from Logger, Metrics,
	// FlightRecorder, and DLQ are appended after these, so derived wiring
	// wins conflicts.
	HostOptions []projectionhost.HostOption
}

// DLQConfig configures the projection dead-letter queue.
type DLQConfig struct {
	// Threshold is the number of handler failures before an event is
	// quarantined. Values <= 0 keep the projectionhost default (3).
	Threshold int

	// Store persists dead-letter entries. When nil (default), a SQLite-backed
	// store is created in the service's own SQLite database (table
	// projection_dead_letters), so entries survive restarts. Required when
	// the resolved driver is not "sqlite". Provide
	// projectionhost.NewMemoryDeadLetterStore for ephemeral tests.
	Store projectionhost.DeadLetterStore
}

// EventService manages a CQRS/ES event store via go-cqrs-lite/system.
// It wraps a system.System and its projection host with lifecycle
// management that integrates with appkit.Service.
type EventService struct {
	sys    *system.System
	dlq    projectionhost.DeadLetterStore
	auxDB  *sql.DB
	inFile *inFlightTracker
	mu     sync.Mutex
	closed bool
}

// NewEventService creates an EventService from the given config.
// The engines declared by the resolved deployment config are opened
// (schema is auto-migrated by the drivers) and the projection host is
// initialized (but not started).
func NewEventService(cfg EventConfig) (*EventService, error) {
	deployment, err := resolveDeployment(cfg)
	if err != nil {
		return nil, err
	}

	aux, dlqStore, cpStore, err := openAuxResources(cfg, deployment)
	if err != nil {
		return nil, err
	}

	sys, inFile, err := buildSystem(cfg, deployment, cpStore, dlqStore)
	if err != nil {
		return nil, closeOnConstructionFailure(aux, err)
	}

	if aux != nil {
		sys.RegisterCloser(auxCloserName, aux)
	}

	return &EventService{ //nolint:exhaustruct_v5 // zero-value mu and closed
		sys:    sys,
		dlq:    dlqStore,
		auxDB:  aux,
		inFile: inFile,
	}, nil
}

// resolveDeployment maps EventConfig onto a system.DeploymentConfig using
// the documented precedence: Deployment > ConfigPath > Driver/DSN/Pragmas.
func resolveDeployment(cfg EventConfig) (system.DeploymentConfig, error) {
	if cfg.Deployment != nil {
		return *cfg.Deployment, nil
	}

	if cfg.ConfigPath != "" {
		loaded, err := system.LoadConfig(cfg.ConfigPath)
		if err != nil {
			return system.DeploymentConfig{}, errorfamily.WrapInfrastructuref(
				err, "cqrs.config_load_failed", "failed to load config %q", cfg.ConfigPath,
			)
		}

		return loaded, nil
	}

	driver, dsn, pragmas, err := resolveStorage(cfg)
	if err != nil {
		return system.DeploymentConfig{}, err
	}

	return defaultDeployment(driver, dsn, pragmas), nil
}

// defaultSQLitePragmas returns the pragmas applied when the resolved
// driver is sqlite and the consumer supplied none. WAL enables concurrent
// readers next to the projection host's checkpoint/DLQ writes; busy_timeout
// absorbs transient lock contention. v0.4.0 (stack/sqlite) shipped the
// same defaults.
func defaultSQLitePragmas() []string {
	return []string{"journal_mode=WAL", "busy_timeout=5000"}
}

// resolveStorage resolves the Driver/DSN/Pragmas triple, honoring the
// deprecated SQLitePath alias. A missing DSN is a Rejection unless the
// driver is explicitly "memory" (in-process store for tests).
func resolveStorage(cfg EventConfig) (string, string, []string, error) {
	driver := cfg.Driver
	dsn := cfg.DSN

	if cfg.SQLitePath != "" {
		dsn = cfg.SQLitePath

		if driver == "" {
			driver = defaultSQLiteDriver
		}
	}

	if driver == "" {
		driver = defaultSQLiteDriver
	}

	if dsn == "" && driver != memoryDriver {
		return "", "", nil, errorfamily.NewRejection(
			"cqrs.path_required",
			"DSN is required (use Driver \"memory\" for an in-process store)",
		)
	}

	pragmas := cfg.Pragmas

	if driver == defaultSQLiteDriver && len(pragmas) == 0 {
		pragmas = defaultSQLitePragmas()
	}

	return driver, dsn, pragmas, nil
}

// defaultDeployment builds the single-engine deployment mirroring the
// reference consumer pattern: one engine, one source-of-truth instance and
// one projections instance.
func defaultDeployment(driver, dsn string, pragmas []string) system.DeploymentConfig {
	return system.DeploymentConfig{ //nolint:exhaustruct_v5 // optional operator fields intentionally zero
		Engines: map[string]system.EngineConfig{
			defaultEngineName: { //nolint:exhaustruct_v5 // priority and views intentionally default
				Driver: driver, DSN: dsn, Pragmas: pragmas,
			},
		},
		Instances: []system.InstanceConfig{
			{ //nolint:exhaustruct_v5 // optional instance fields intentionally zero
				Role: system.RoleSourceOfTruth, Engine: defaultEngineName,
			},
			{ //nolint:exhaustruct_v5 // optional instance fields intentionally zero
				Role: system.RoleProjections, Engine: defaultEngineName,
			},
		},
	}
}

// openAuxResources opens the auxiliary *sql.DB backing the default
// persistent checkpoint and DLQ stores. It returns no aux handle when the
// deployment is not sqlite-with-file (or both stores are consumer-supplied
// or absent).
func openAuxResources( //nolint:ireturn // upstream interface
	cfg EventConfig,
	deployment system.DeploymentConfig,
) (*sql.DB, projectionhost.DeadLetterStore, event.CheckpointStore, error) {
	wantDefaultDLQ := wantsDefaultDLQ(cfg)
	wantDefaultCP := cfg.CheckpointStore == nil
	sqliteFile := deploymentUsesSQLiteFile(deployment)

	if (!wantDefaultDLQ && !wantDefaultCP) || !sqliteFile {
		if wantDefaultDLQ && !sqliteFile {
			return nil, nil, nil, errorfamily.NewRejection(
				"cqrs.dlq_store_required",
				`DLQConfig.Store is required when the driver is not "sqlite"`,
			)
		}

		return nil, dlqStoreOrNil(cfg), cfg.CheckpointStore, nil
	}

	aux, openErr := sql.Open(defaultSQLiteDriver, auxDSN(cfg, deployment))
	if openErr != nil {
		return nil, nil, nil, errorfamily.WrapInfrastructuref(
			openErr,
			"cqrs.open_failed",
			"failed to open auxiliary database at %s",
			auxDSN(cfg, deployment),
		)
	}

	dlqStore, cpStore, storeErr := buildAuxStores(context.Background(), cfg, aux)
	if storeErr != nil {
		_ = aux.Close()

		return nil, nil, nil, storeErr
	}

	return aux, dlqStore, cpStore, nil
}

// deploymentUsesSQLiteFile reports whether the deployment resolves to a
// sqlite driver with a file DSN (the only shape the aux handle supports).
func deploymentUsesSQLiteFile(deployment system.DeploymentConfig) bool {
	for _, eng := range deployment.Engines {
		if eng.Driver == defaultSQLiteDriver && eng.DSN != "" {
			return true
		}
	}

	return false
}

// wantsDefaultDLQ reports whether the config needs a default (SQL-backed)
// dead-letter store.
func wantsDefaultDLQ(cfg EventConfig) bool {
	return cfg.DLQ != nil && cfg.DLQ.Store == nil
}

// dlqStoreOrNil returns the consumer-supplied dead-letter store, if any.
func dlqStoreOrNil(cfg EventConfig) projectionhost.DeadLetterStore { //nolint:ireturn // upstream interface
	if cfg.DLQ == nil {
		return nil
	}

	return cfg.DLQ.Store
}

// auxDSN picks the file DSN the aux handle opens: the first sqlite file
// engine in the deployment, falling back to the config-level DSN. A busy
// timeout is injected so checkpoint/DLQ writes wait out engine-side lock
// contention instead of failing with SQLITE_BUSY.
func auxDSN(cfg EventConfig, deployment system.DeploymentConfig) string {
	dsn := ""

	for _, name := range sortedEngineNames(deployment) {
		eng := deployment.Engines[name]

		if eng.Driver == defaultSQLiteDriver && eng.DSN != "" {
			dsn = eng.DSN

			break
		}
	}

	if dsn == "" {
		if cfg.SQLitePath != "" {
			dsn = cfg.SQLitePath
		} else {
			dsn = cfg.DSN
		}
	}

	const busyParam = "_pragma=busy_timeout(5000)"

	if strings.Contains(dsn, "?") {
		return dsn + "&" + busyParam
	}

	return dsn + "?" + busyParam
}

// sortedEngineNames returns the deployment's engine names in deterministic
// order.
func sortedEngineNames(deployment system.DeploymentConfig) []string {
	names := make([]string, 0, len(deployment.Engines))
	for name := range deployment.Engines {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// buildAuxStores creates the default DLQ and checkpoint stores on the aux
// handle, honoring consumer overrides.
func buildAuxStores( //nolint:ireturn // upstream interface
	ctx context.Context,
	cfg EventConfig,
	handle *sql.DB,
) (projectionhost.DeadLetterStore, event.CheckpointStore, error) {
	dlqStore := dlqStoreOrNil(cfg)

	if wantsDefaultDLQ(cfg) {
		store, err := projectionhost.NewSQLiteDeadLetterStore(ctx, handle)
		if err != nil {
			return nil, nil, errorfamily.WrapInfrastructure(
				err, "cqrs.dlq_provision_failed", "failed to create dead-letter store",
			)
		}

		dlqStore = store
	}

	cpStore := cfg.CheckpointStore

	if cpStore == nil {
		schemaErr := applyCheckpointSchema(ctx, handle)
		if schemaErr != nil {
			return nil, nil, schemaErr
		}

		store, err := eventstore.NewSQLiteCheckpointStore(handle)
		if err != nil {
			return nil, nil, errorfamily.WrapInfrastructure(
				err, "cqrs.checkpoint_provision_failed", "failed to create checkpoint store",
			)
		}

		cpStore = store
	}

	return dlqStore, cpStore, nil
}

// applyCheckpointSchema creates the checkpoint table on the aux handle.
func applyCheckpointSchema(ctx context.Context, handle *sql.DB) error {
	_, err := handle.ExecContext(ctx, eventstore.SQLiteCheckpointSchema())
	if err != nil {
		return errorfamily.WrapInfrastructure(
			err, "cqrs.checkpoint_provision_failed", "failed to create checkpoint schema",
		)
	}

	return nil
}

// hostBootstrapDeclaration names the zero-entry count projection that
// guarantees system.New creates the projection host even when consumers only
// register raw host projections (the host is built exclusively when
// DomainConfig.Projections is non-empty). It counters no event types and has
// no read-model cost.
const hostBootstrapDeclaration = "appkit-host"

// bootstrapSample is the decoder sample for the never-emitted bootstrap
// event type.
type bootstrapSample struct {
	ID string
}

// buildSystem constructs the system.System with derived host options and
// middleware wiring for the C/Q facade. The in-flight drain tracker is
// installed outermost so Shutdown waits for entire command chains. dlqStore
// is the resolved default dead-letter store (nil when DLQ is disabled or
// consumer-supplied).
func buildSystem(
	cfg EventConfig,
	deployment system.DeploymentConfig,
	cpStore event.CheckpointStore,
	dlqStore projectionhost.DeadLetterStore,
) (*system.System, *inFlightTracker, error) {
	inFile := newInFlightTracker()

	middleware := append([]command.Middleware{inFile.commandMiddleware()}, cfg.CommandMiddleware...)

	domain := system.DomainConfig{ //nolint:exhaustruct_v5 // domain registration is the consumer's job
		Middleware: middleware,
		Projections: []system.ProjectionDeclaration{
			// Host bootstrap: a count over an event type no domain emits, so
			// system.New always creates the projection host even when
			// consumers only register raw host projections.
			system.Count(hostBootstrapDeclaration).
				On("appkit.internal.never", bootstrapSample{ID: ""}, 1, "ID").
				Done(),
		},
		ProjectionHostOptions: cfg.hostOptions(dlqStore),
		CheckpointStore:       cpStore,
	}

	sys, err := system.New(context.Background(), domain, deployment)
	if err != nil {
		return nil, nil, errorfamily.WrapInfrastructuref(
			err, "cqrs.system_failed", "failed to create CQRS system",
		)
	}

	if len(cfg.QueryMiddleware) > 0 {
		sys.QueryDispatcher().Use(cfg.QueryMiddleware...)
	}

	return sys, inFile, nil
}

// closeOnConstructionFailure tears down the half-built aux handle when
// NewEventService aborts. The primary error is returned untouched when the
// close succeeds; a close failure is appended with errors.Join so a double
// failure is never silently discarded.
func closeOnConstructionFailure(aux io.Closer, err error) error {
	if aux == nil {
		return err
	}

	if db, ok := aux.(*sql.DB); ok && db == nil {
		return err
	}

	closeErr := aux.Close()
	if closeErr != nil {
		return errors.Join(err, closeErr)
	}

	return err
}

// hostOptions maps EventConfig onto projectionhost options.
// Nil-valued config fields are skipped so projectionhost defaults apply.
// Consumer-supplied HostOptions come first; derived wiring wins conflicts.
// dlqStore is the default (SQL-backed) store resolved at construction; a
// consumer-supplied DLQConfig.Store takes precedence.
func (cfg EventConfig) hostOptions(
	dlqStore projectionhost.DeadLetterStore,
) []projectionhost.HostOption {
	opts := append([]projectionhost.HostOption{}, cfg.HostOptions...)

	if cfg.Logger != nil {
		opts = append(opts, projectionhost.WithLogger(cfg.Logger))
	}

	if cfg.FlightRecorder != nil {
		opts = append(opts,
			projectionhost.WithFlightRecorder(cfg.FlightRecorder, cfg.FlightRecorderTrigger))
	}

	if cfg.Metrics != nil {
		opts = append(opts, projectionhost.WithMetrics(cfg.Metrics))
	}

	if cfg.DLQ != nil {
		store := cfg.DLQ.Store
		if store == nil {
			store = dlqStore
		}

		if store != nil {
			opts = append(opts,
				projectionhost.WithDeadLetterStore(store, cfg.DLQ.Threshold))
		}
	}

	return opts
}

// System returns the underlying system.System.
// Use it for advanced wiring: MetaEngine, Publisher, EventStore,
// SnapshotStore, ProjectionPlan, introspection, and more.
func (es *EventService) System() *system.System {
	return es.sys
}

// Host returns the underlying projectionhost.Host, or nil for deployments
// without a RoleProjections instance.
// Use it to register projections before starting the service.
func (es *EventService) Host() *projectionhost.Host {
	return es.sys.ProjectionHost()
}

// DeadLetterStore returns the configured dead-letter store, or nil when the
// DLQ is disabled. The SQLite default additionally implements
// projectionhost.DeadLetterStoreAdmin (Count, ListPaged, PurgeBefore) via
// type assertion.
func (es *EventService) DeadLetterStore() projectionhost.DeadLetterStore { //nolint:ireturn // upstream interface
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

	return es.host().ReplayDeadLetters(ctx, projectionName) //nolint:wrapcheck // delegation
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
	return es.host().Reset(ctx, name, opts...) //nolint:wrapcheck // delegation
}

// DB returns the auxiliary *sql.DB backing the default checkpoint and DLQ
// stores. Returns a Rejection when no auxiliary database exists (memory
// deployments, or fully consumer-supplied stores).
func (es *EventService) DB() (*sql.DB, error) {
	if es.auxDB == nil {
		return nil, errorfamily.NewRejection(
			"cqrs.db_not_sql", "no auxiliary SQL database for this deployment",
		)
	}

	return es.auxDB, nil
}

// ReadyCheck reports whether all registered projections are serving: every
// worker must be live (caught up and processing) or stopped (fully drained,
// normal for batch-style hosts). A worker that is idle before
// StartProjections, still catching up, backing off after a crash, draining,
// or terminally failed makes the service NOT ready — a not-yet-started
// service is never ready. Wire it into appkit.ServiceConfig.ReadyCheck so
// /health/ready serves 503 until projections are caught up and flips back if
// one dies:
//
//	cfg.ReadyCheck = eventSvc.ReadyCheck
//
// The service always runs the system auto-projection worker (the metaengine
// read models); its presence after StartProjections does not block readiness.
func (es *EventService) ReadyCheck() bool {
	for _, state := range es.host().Status() {
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
	return es.host().LagPerProjection()
}

// CheckStaleness reports a Transient error when the maximum projection lag
// exceeds maxStaleness, nil otherwise. Projections run asynchronously, so a
// command handler returning does not imply the read model has moved — use this
// as a read-time guard before serving data from a read model (or serve 503):
// with lag beyond the threshold, the projection will catch up once its worker
// drains the backlog. A maxStaleness <= 0 disables the check; a projection
// that has not processed any event yet counts as fresh.
func (es *EventService) CheckStaleness(maxStaleness time.Duration) error {
	return es.host().CheckStaleness(maxStaleness) //nolint:wrapcheck // delegation
}

// CheckProjectionStaleness is the per-projection variant of CheckStaleness:
// it guards reads served from one named read model instead of the maximum
// across all workers. Rejects (400-class) when the projection is not
// registered; a maxStaleness <= 0 disables the check first.
func (es *EventService) CheckProjectionStaleness(name string, maxStaleness time.Duration) error {
	return es.host().CheckProjectionStaleness(name, maxStaleness) //nolint:wrapcheck // delegation
}

// StartProjections starts the projection host workers.
// Must be called after all projections are registered and before the service begins serving.
func (es *EventService) StartProjections(ctx context.Context) error {
	return es.sys.Start(ctx) //nolint:wrapcheck // delegation
}

// Shutdown gracefully stops projections and closes the event store.
// In-flight commands are drained first (bounded by the context), then the
// system closes in dependency order. Safe to call multiple times (idempotent
// via mutex guard).
func (es *EventService) Shutdown(ctx context.Context) error {
	es.mu.Lock()

	if es.closed {
		es.mu.Unlock()

		return nil
	}

	es.closed = true
	es.mu.Unlock()

	drainErr := es.inFile.drain(ctx)
	if drainErr != nil {
		return fmt.Errorf("cqrs: drain in-flight commands: %w", drainErr)
	}

	return es.sys.GracefulClose(ctx) //nolint:wrapcheck // delegation
}

// host returns the projection host, tolerating deployments without one.
func (es *EventService) host() *projectionhost.Host {
	return es.sys.ProjectionHost()
}
