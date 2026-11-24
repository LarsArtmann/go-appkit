package cqrs

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	cqrsotel "github.com/larsartmann/go-cqrs-lite/otel/v4"
	"github.com/larsartmann/go-cqrs-lite/query/v4"
)

// The audit middlewares persist every dispatched command/query into the
// system's audit stores. go-cqrs-lite ships CommandStore()/QueryStore() as
// passive sinks — nothing upstream writes to them, so the cqrs-htmx
// /cqrs/commands and /cqrs/queries panels stay permanently empty unless the
// composition root records dispatches itself (go-appkit#83). Both recorders
// are default-ON (EventConfig.Audit) and best-effort: a persistence failure
// is logged and never fails the business operation.
//
// Tracing metadata is stamped into every persisted record: the source
// command's metadata is propagated, correlation and request IDs are minted
// when absent, and an OTel-baggage correlation ID (cqrsotel) wins when the
// context carries one. Causation is never fabricated.

// commandAuditor records command dispatches into the system's command audit
// store. The store only exists after system.New builds it, so the middleware
// is inert until bind is called (the middleware list is assembled first).
type commandAuditor struct {
	store  atomic.Pointer[command.Store]
	logger *slog.Logger
}

func newCommandAuditor(logger *slog.Logger) *commandAuditor {
	if logger == nil {
		logger = slog.Default()
	}

	return &commandAuditor{logger: logger}
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
		return command.StreamRef{}, false
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
	out := command.Metadata{}
	if carrier, ok := cmd.(interface{ Metadata() command.Metadata }); ok {
		out = carrier.Metadata()
	}

	if raw := cqrsotel.CorrelationIDFromContext(ctx); raw != "" {
		if parsed, err := id.ParseCorrelationID(raw); err == nil {
			out.CorrelationID = parsed
		}
	}

	if out.CorrelationID.IsZero() {
		out.CorrelationID = id.NewCorrelationID()
	}

	if out.RequestID.IsZero() {
		out.RequestID = id.NewRequestID()
	}

	return out
}

// queryAuditor records query dispatches into the system's query audit store,
// type + request ID + timing only (no payload — read-side records stay small).
type queryAuditor struct {
	store  atomic.Pointer[query.QueryStore]
	logger *slog.Logger
}

func newQueryAuditor(logger *slog.Logger) *queryAuditor {
	if logger == nil {
		logger = slog.Default()
	}

	return &queryAuditor{logger: logger}
}

func (a *queryAuditor) bind(store query.QueryStore) {
	if store == nil {
		return
	}

	a.store.Store(&store)
}

func (a *queryAuditor) middleware() query.Middleware {
	return func(next query.Handler) query.Handler {
		return func(ctx context.Context, q query.Query) (any, error) {
			start := time.Now()

			result, err := next(ctx, q)

			a.record(ctx, q, start)

			return result, err
		}
	}
}

func (a *queryAuditor) record(ctx context.Context, q query.Query, start time.Time) {
	store := a.store.Load()
	if store == nil || q == nil {
		return
	}

	opts := []query.QueryPersistOption{
		query.WithQueryReceivedAt(start),
	}

	if carrier, ok := q.(query.MetadataCarrier); ok {
		md := carrier.Metadata()
		if !md.RequestID.IsZero() {
			opts = append(opts, query.WithQueryID(md.RequestID))
		}

		if md.CorrelationID.IsZero() {
			if raw := cqrsotel.CorrelationIDFromContext(ctx); raw != "" {
				if parsed, err := id.ParseCorrelationID(raw); err == nil {
					md.CorrelationID = parsed
				}
			}
		}

		if md.CorrelationID.IsZero() {
			md.CorrelationID = id.NewCorrelationID()
		}

		opts = append(opts, query.WithQueryMetadata(md))
	}

	persisted, err := query.NewPersistedQuery(q.Type(), nil, opts...)
	if err != nil {
		a.logger.WarnContext(ctx, "cqrs query audit: build persisted query failed",
			slog.String("query_type", string(q.Type())), slog.String("error", err.Error()))

		return
	}

	// The record must survive the request lifecycle: a client giving up
	// between handler completion and this save is exactly the query most
	// worth auditing.
	if saveErr := (*store).SaveQuery(context.WithoutCancel(ctx), persisted); saveErr != nil {
		a.logger.WarnContext(ctx, "cqrs query audit: save failed",
			slog.String("query_type", string(q.Type())), slog.String("error", saveErr.Error()))
	}
}
