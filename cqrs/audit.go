package cqrs

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	cqrsotel "github.com/larsartmann/go-cqrs-lite/otel/v4"
	"github.com/larsartmann/go-cqrs-lite/query/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
)

// The audit trail persists every dispatched command and query into the
// system's audit stores. go-cqrs-lite ships CommandStore()/QueryStore() as
// passive sinks: a bare system.New leaves them unwritten, so the cqrs-htmx
// /cqrs/commands and /cqrs/queries panels stay permanently empty unless the
// composition root records dispatches itself (go-appkit#83). Both recorders
// are default-ON (EventConfig.Audit) and best-effort: a persistence failure
// is logged and never fails the business operation.
//
// The two sides are wired differently, because go-cqrs-lite ships only one of
// the two writers. The query side installs the native query.AuditMiddleware
// (one writer for the query log everywhere; EventConfig.AuditPayload selects
// its level). The command side is a local recorder, because no native
// command audit middleware exists yet.
//
// Tracing metadata is stamped into every persisted record: the source
// record's metadata is propagated and request IDs are minted when absent.
// A correlation ID is resolved per side: on a command the OTel-baggage
// correlation ID (cqrsotel) wins over the metadata and a missing one is
// minted; on a query the query's own correlation ID wins, the OTel one is
// stamped onto a query that carries none, and one is minted as the last
// resort. Causation is never fabricated.

// commandAuditor records command dispatches into the system's command audit
// store. The store only exists after system.New builds it, so the middleware
// is inert until bind is called (the middleware list is assembled first).
type commandAuditor struct {
	store  atomic.Pointer[command.Store]
	logger *slog.Logger
}

func newCommandAuditor(logger *slog.Logger) *commandAuditor {
	//art-dupl:accept nil-logger fallback to slog.Default is a cross-module idiom (otel/diagnostics.go); sharing would couple two independently versioned modules
	if logger == nil {
		logger = slog.Default()
	}

	out := new(commandAuditor)
	out.logger = logger

	return out
}

// bind attaches the audit sink once the system has constructed it. A nil
// store keeps the auditor inert (deployments without a command store).
func (a *commandAuditor) bind(store command.Store) {
	if store == nil {
		return
	}

	a.store.Store(&store)
}

// middleware returns the chain entry; install it directly after the in-flight
// tracker so commands are audited even when downstream middleware rejects.
func (a *commandAuditor) middleware() command.Middleware {
	return func(next command.Handler) command.Handler {
		return func(ctx context.Context, cmd command.Command) error {
			a.record(ctx, cmd)

			return next(ctx, cmd)
		}
	}
}

func (a *commandAuditor) record(ctx context.Context, cmd command.Command) {
	store := a.store.Load()
	if store == nil || cmd == nil {
		return
	}

	ref, ok := auditStreamRef(cmd)
	if !ok {
		return
	}

	persisted, err := command.NewPersistedCommand(
		cmd.Type(), ref, nil,
		command.WithPersistedCommandID(cmd.ID()),
		command.WithCommandMetadata(commandAuditMetadata(ctx, cmd)),
	)
	if err != nil {
		a.logger.WarnContext(ctx, "cqrs command audit: build persisted command failed",
			slog.String("command_type", string(cmd.Type())), slog.String("error", err.Error()))

		return
	}

	if saveErr := (*store).Save(ctx, ref, persisted); saveErr != nil {
		a.logger.WarnContext(ctx, "cqrs command audit: save failed",
			slog.String("command_type", string(cmd.Type())), slog.String("error", saveErr.Error()))
	}
}

// auditStreamRef builds the store reference for a dispatched command. v4
// commands carry no stream type, so the audit trail buckets under
// "appkit" unless the command exposes its own StreamRef or StreamType
// (PersistedCommand re-dispatch and future v5-typed commands do).
func auditStreamRef(cmd command.Command) (command.StreamRef, bool) {
	if carrier, ok := cmd.(interface{ StreamRef() command.StreamRef }); ok {
		return carrier.StreamRef(), true
	}

	if carrier, ok := cmd.(interface{ StreamType() command.StreamType }); ok {
		return command.StreamRef{Type: carrier.StreamType(), ID: cmd.StreamID()}, true
	}

	ref := command.StreamRef{Type: auditStreamBucket, ID: cmd.StreamID()}
	if ref.ID.IsZero() {
		var zero command.StreamRef

		return zero, false
	}

	return ref, true
}

// auditStreamBucket is the audit-trail stream type for commands that do not
// carry their own (v4 BasicCommand has no stream type — the record still
// carries the true stream ID and command type).
const auditStreamBucket = command.StreamType("appkit")

// commandAuditMetadata returns the command's metadata with tracing guaranteed:
// an OTel-context correlation ID wins, existing metadata values are
// preserved, and correlation/request IDs are minted when still absent.
func commandAuditMetadata(ctx context.Context, cmd command.Command) command.Metadata {
	var out command.Metadata
	if carrier, ok := cmd.(interface{ Metadata() command.Metadata }); ok {
		out = carrier.Metadata()
	}

	if correlationID, ok := correlationIDFromContext(ctx); ok {
		out.CorrelationID = correlationID
	}

	if out.CorrelationID.IsZero() {
		out.CorrelationID = id.NewCorrelationID()
	}

	if out.RequestID.IsZero() {
		out.RequestID = id.NewRequestID()
	}

	return out
}

// installQueryAudit records every dispatched query into the system's query
// audit store through go-cqrs-lite's native query.AuditMiddleware: one
// writer for the query log, here and in every other go-cqrs-lite consumer.
// The audit middleware is outermost, so a query is recorded even when
// consumer middleware rejects it; the correlation guard runs directly inside
// it and stamps a correlation ID onto the query before anything else sees it,
// so the record always correlates with the request. A deployment without a
// query store installs nothing.
func installQueryAudit(sys *system.System, level query.AuditLevel, logger *slog.Logger) {
	store := sys.QueryStore()
	if store == nil {
		return
	}

	sys.QueryDispatcher().Use(query.AuditMiddleware(store, level, logger), ensureQueryCorrelation())
}

// queryAuditLevel maps the config onto the native audit level: metadata only
// by default, payload included when EventConfig.AuditPayload opts in.
func queryAuditLevel(cfg EventConfig) query.AuditLevel {
	if cfg.AuditPayload != nil && *cfg.AuditPayload {
		return query.AuditFull
	}

	return query.AuditMetadata
}

// ensureQueryCorrelation is installed inside the audit middleware. The
// native middleware carries a query's own metadata verbatim, so this guard
// guarantees the one field an audit record must never lack: the correlation
// ID. An OTel baggage correlation ID wins; otherwise one is minted.
func ensureQueryCorrelation() query.Middleware {
	return func(next query.Handler) query.Handler {
		return func(ctx context.Context, q query.Query) (any, error) {
			stampQueryCorrelation(ctx, q)

			return next(ctx, q)
		}
	}
}

// stampQueryCorrelation sets a correlation ID on a query that carries none.
// Only a query that applies options in place (the BasicQuery shape every
// transport decoder builds) can be stamped; any other shape keeps its own
// metadata.
func stampQueryCorrelation(ctx context.Context, qry query.Query) {
	carrier, ok := qry.(query.MetadataCarrier)
	if !ok || !carrier.Metadata().CorrelationID.IsZero() {
		return
	}

	applier, ok := qry.(interface{ ApplyOptions(opts ...query.Option) })
	if !ok {
		return
	}

	applier.ApplyOptions(query.WithCorrelationID(auditCorrelationID(ctx)))
}

// correlationIDFromContext returns the correlation ID the OTel context
// carries. False when the context carries none or a value that is not a
// correlation ID.
func correlationIDFromContext(ctx context.Context) (id.CorrelationID, bool) {
	raw := cqrsotel.CorrelationIDFromContext(ctx)
	if raw == "" {
		return id.CorrelationID{}, false
	}

	parsed, err := id.ParseCorrelationID(raw)
	if err != nil {
		return id.CorrelationID{}, false
	}

	return parsed, true
}

// auditCorrelationID resolves the correlation ID for an audit record: the
// OTel context's when it carries a valid one, else a fresh ID.
func auditCorrelationID(ctx context.Context) id.CorrelationID {
	if correlationID, ok := correlationIDFromContext(ctx); ok {
		return correlationID
	}

	return id.NewCorrelationID()
}
