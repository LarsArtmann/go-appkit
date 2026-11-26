package cqrs

// The audit tests pin the default-ON audit trail (go-appkit#83): every
// dispatched command and query persists into the system's audit stores with
// tracing metadata stamped (propagated, OTel-baggage-correlated, or minted),
// the EventConfig.Audit opt-out empties the trail, and persistence failures
// never break dispatch.

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	cqrsotel "github.com/larsartmann/go-cqrs-lite/otel/v4"
	"github.com/larsartmann/go-cqrs-lite/query/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
	errorfamily "github.com/larsartmann/go-error-family"
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

// auditedQuery is the BasicQuery shape every transport decoder builds, plus
// the payload capability the native audit middleware reads at AuditFull.
type auditedQuery struct {
	*query.BasicQuery

	body []byte
}

func (q *auditedQuery) Payload() []byte { return slices.Clone(q.body) }

func newAuditedQuery(t *testing.T, body []byte) *auditedQuery {
	t.Helper()

	basic, err := query.New("audit.search")
	if err != nil {
		t.Fatalf("query.New: %v", err)
	}

	return &auditedQuery{BasicQuery: basic, body: body} //nolint:exhaustruct_v5 // test fixture: only the payload field is optional
}

func registerAuditedQuery(t *testing.T, eventSvc *EventService) {
	t.Helper()

	err := RegisterQuery(eventSvc, "audit.search",
		func(_ context.Context, _ *auditedQuery) (int, error) {
			return 1, nil
		})
	if err != nil {
		t.Fatalf("RegisterQuery: %v", err)
	}
}

func auditedSearchRecords(t *testing.T, eventSvc *EventService) []*query.PersistedQuery {
	t.Helper()

	records, err := eventSvc.QueryStore().LoadQueries(context.Background(), time.Time{})
	if err != nil {
		t.Fatalf("load query audit records: %v", err)
	}

	var found []*query.PersistedQuery

	for _, record := range records {
		if record.Type() == "audit.search" {
			found = append(found, record)
		}
	}

	return found
}

// TestEventService_AuditQueryCarriesIdentity pins the user mapping: a query
// stamped the way every HTTP transport stamps it is recorded with its user,
// its request ID, and a correlation ID.
func TestEventService_AuditQueryCarriesIdentity(t *testing.T) {
	t.Parallel()

	eventSvc := newFacadeService(t)
	registerAuditedQuery(t, eventSvc)

	userID, requestID := id.NewUserID(), id.NewRequestID()
	qry := newAuditedQuery(t, []byte(`{"q":"shoes"}`))
	qry.ApplyOptions(query.WithUserID(userID), query.WithRequestID(requestID))

	if _, err := DispatchQuery[*auditedQuery, int](context.Background(), eventSvc, qry); err != nil {
		t.Fatalf("dispatch query: %v", err)
	}

	records := auditedSearchRecords(t, eventSvc)
	if len(records) != 1 {
		t.Fatalf("audit records: got %d, want 1", len(records))
	}

	meta := records[0].Metadata()

	if meta.UserID != userID {
		t.Errorf("recorded user = %v, want %v", meta.UserID, userID)
	}

	if records[0].ID() != requestID {
		t.Errorf("recorded request ID = %v, want %v", records[0].ID(), requestID)
	}

	if meta.CorrelationID.IsZero() {
		t.Error("correlation ID is zero: the audit guard must stamp one")
	}
}

// TestEventService_AuditQueryPayloadOptIn pins the privacy default: the
// payload (what a user typed) is recorded only when AuditPayload opts in.
func TestEventService_AuditQueryPayloadOptIn(t *testing.T) {
	t.Parallel()

	enabled := true
	disabled := false
	body := []byte(`{"q":"shoes"}`)

	cases := []struct {
		name        string
		payload     *bool
		wantPayload bool
	}{
		{name: "default omits the payload", payload: nil, wantPayload: false},
		{name: "explicit false omits the payload", payload: &disabled, wantPayload: false},
		{name: "opt-in records the payload", payload: &enabled, wantPayload: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := EventConfig{Driver: memoryDriver, AuditPayload: tc.payload}
			eventSvc := newFacadeServiceCfg(t, cfg)
			registerAuditedQuery(t, eventSvc)

			qry := newAuditedQuery(t, body)
			if _, err := DispatchQuery[*auditedQuery, int](context.Background(), eventSvc, qry); err != nil {
				t.Fatalf("dispatch query: %v", err)
			}

			records := auditedSearchRecords(t, eventSvc)
			if len(records) != 1 {
				t.Fatalf("audit records: got %d, want 1", len(records))
			}

			got := records[0].Payload()

			if tc.wantPayload && string(got) != string(body) {
				t.Errorf("recorded payload = %q, want %q", got, body)
			}

			if !tc.wantPayload && got != nil {
				t.Errorf("recorded payload = %q, want none", got)
			}
		})
	}
}

// TestEventService_AuditQuerySurvivesConsumerRejection pins the install
// order: the audit middleware sits outside consumer middleware, so a query
// is recorded even when that middleware rejects it.
func TestEventService_AuditQuerySurvivesConsumerRejection(t *testing.T) {
	t.Parallel()

	eventSvc := newFacadeService(t)
	registerAuditedQuery(t, eventSvc)

	eventSvc.QueryDispatcher().Use(func(query.Handler) query.Handler {
		return func(context.Context, query.Query) (any, error) {
			return nil, errorfamily.NewRejection("test.rejected", "rejected by consumer middleware")
		}
	})

	qry := newAuditedQuery(t, nil)
	if _, err := DispatchQuery[*auditedQuery, int](context.Background(), eventSvc, qry); err == nil {
		t.Fatal("dispatch succeeded, want the consumer middleware's rejection")
	}

	records := auditedSearchRecords(t, eventSvc)
	if len(records) != 1 {
		t.Fatalf("audit records after a rejection: got %d, want 1", len(records))
	}
}

// TestEventService_AuditQueryWithoutStore pins the nil guard: a deployment
// that builds no query store records nothing and still answers queries.
func TestEventService_AuditQueryWithoutStore(t *testing.T) {
	t.Parallel()

	projectionOnly := system.DeploymentConfig{ //nolint:exhaustruct_v5 // projection-only deployment: the query-audit role is deliberately absent
		Engines: map[string]system.EngineConfig{
			defaultEngineName: {Driver: memoryDriver},
		},
		Instances: []system.InstanceConfig{{Role: system.RoleProjections, Engine: defaultEngineName}},
	}

	cfg := EventConfig{ //nolint:exhaustruct_v5 // deployment under test: Driver and Deployment are the only inputs
		Driver:     memoryDriver,
		Deployment: &projectionOnly,
	}

	eventSvc, err := NewEventService(cfg)
	if err != nil {
		t.Fatalf("NewEventService: %v", err)
	}

	t.Cleanup(func() { _ = eventSvc.Shutdown(context.Background()) })

	if eventSvc.QueryStore() != nil {
		t.Fatalf("QueryStore = %T, want nil for a deployment without a query-audit role", eventSvc.QueryStore())
	}

	registerAuditedQuery(t, eventSvc)

	result, err := DispatchQuery[*auditedQuery, int](context.Background(), eventSvc, newAuditedQuery(t, nil))
	if err != nil {
		t.Fatalf("dispatch query without an audit store: %v", err)
	}

	if result != 1 {
		t.Errorf("result = %d, want 1", result)
	}
}
