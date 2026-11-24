package cqrs

// The audit tests pin the default-ON audit trail (go-appkit#83): every
// dispatched command and query persists into the system's audit stores with
// tracing metadata stamped (propagated, OTel-baggage-correlated, or minted),
// the EventConfig.Audit opt-out empties the trail, and persistence failures
// never break dispatch.

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	cqrsotel "github.com/larsartmann/go-cqrs-lite/otel/v4"
)

func TestEventService_AuditPersistsCommandWithTracing(t *testing.T) {
	t.Parallel()

	eventSvc := newFacadeService(t)

	streamID := id.NewStreamID()
	cmd := newFacadeCommand(t, streamID)

	if err := eventSvc.Dispatch(context.Background(), cmd); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	store := eventSvc.CommandStore()
	if store == nil {
		t.Fatal("CommandStore is nil — the audit store must be exposed")
	}

	ref := command.StreamRef{Type: auditStreamBucket, ID: streamID}

	records, err := store.Load(context.Background(), ref)
	if err != nil {
		t.Fatalf("load audit records: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("audit records: got %d, want 1", len(records))
	}

	meta := records[0].Metadata()
	if meta.CorrelationID.IsZero() {
		t.Error("correlation ID is zero — the auditor must stamp one")
	}

	if meta.RequestID.IsZero() {
		t.Error("request ID is zero — the auditor must stamp one")
	}

	if records[0].ID() != cmd.ID() {
		t.Errorf("record ID = %v, want the dispatched command's %v", records[0].ID(), cmd.ID())
	}
}

func TestEventService_AuditOTelCorrelationWins(t *testing.T) {
	t.Parallel()

	eventSvc := newFacadeService(t)

	fixed := id.NewCorrelationID()
	ctx := cqrsotel.WithCorrelationID(context.Background(), fixed.String())

	cmd := newFacadeCommand(t, id.NewStreamID())
	if err := eventSvc.Dispatch(ctx, cmd); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	ref := command.StreamRef{Type: auditStreamBucket, ID: cmd.StreamID()}

	records, err := eventSvc.CommandStore().Load(context.Background(), ref)
	if err != nil {
		t.Fatalf("load audit records: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("audit records: got %d, want 1", len(records))
	}

	if got := records[0].Metadata().CorrelationID; got != fixed {
		t.Errorf("correlation = %v, want the OTel-context %v", got, fixed)
	}
}

func TestEventService_AuditOptOut(t *testing.T) {
	t.Parallel()

	off := false
	eventSvc := newFacadeServiceCfg(t, EventConfig{Driver: memoryDriver, Audit: &off})

	cmd := newFacadeCommand(t, id.NewStreamID())
	if err := eventSvc.Dispatch(context.Background(), cmd); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	ref := command.StreamRef{Type: auditStreamBucket, ID: cmd.StreamID()}

	records, err := eventSvc.CommandStore().Load(context.Background(), ref)
	if err != nil {
		t.Fatalf("load audit records: %v", err)
	}

	if len(records) != 0 {
		t.Fatalf("opted-out service persisted %d records, want 0", len(records))
	}
}

func TestEventService_AuditPersistsQueries(t *testing.T) {
	t.Parallel()

	eventSvc := newFacadeService(t)

	if _, err := DispatchQuery[facadeQuery, int](context.Background(), eventSvc, facadeQuery{}); err != nil {
		t.Fatalf("dispatch query: %v", err)
	}

	store := eventSvc.QueryStore()
	if store == nil {
		t.Fatal("QueryStore is nil — the audit store must be exposed")
	}

	records, err := store.LoadQueries(context.Background(), time.Time{})
	if err != nil {
		t.Fatalf("load query audit records: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("query audit records: got %d, want 1", len(records))
	}

	if records[0].Type() != "facade.count" {
		t.Errorf("recorded type = %v, want facade.count", records[0].Type())
	}
}
