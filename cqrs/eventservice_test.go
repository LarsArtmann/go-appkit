package cqrs

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
)

func TestNewEventService_EmptyConfigRejected(t *testing.T) {
	t.Parallel()

	_, err := NewEventService(EventConfig{})
	if err == nil {
		t.Fatal("expected error for empty config (no DSN)")
	}

	familyErr, ok := errors.AsType[*errorfamily.Error](err)
	if !ok {
		t.Fatalf("expected *errorfamily.Error, got %T", err)
	}

	if familyErr.Code() != "cqrs.path_required" {
		t.Errorf("expected code cqrs.path_required, got %q", familyErr.Code())
	}
}

func TestNewEventService_FileBackedSQLite(t *testing.T) {
	t.Parallel()

	eventSvc := newTestEventService(t, EventConfig{})
	if eventSvc.System() == nil {
		t.Fatal("expected non-nil System")
	}

	if eventSvc.Host() == nil {
		t.Fatal("expected non-nil Host")
	}

	if eventSvc.System().EventStore() == nil {
		t.Error("expected non-nil EventStore")
	}

	if eventSvc.System().Publisher() == nil {
		t.Error("expected non-nil Publisher")
	}

	if eventSvc.System().MetaEngine() == nil {
		t.Error("expected non-nil MetaEngine")
	}
}

func TestNewEventService_MemoryDriver(t *testing.T) {
	t.Parallel()

	eventSvc, err := NewEventService(EventConfig{
		Driver: "memory",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	_, err = eventSvc.DB()
	if err == nil {
		t.Error("expected DB() rejection for memory deployment (no aux database)")
	}
}

func TestNewEventService_DeploymentOverride(t *testing.T) {
	t.Parallel()

	eventSvc, err := NewEventService(EventConfig{
		Deployment: memoryDeployment(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	if eventSvc.Host() == nil {
		t.Fatal("expected projection host from RoleProjections instance")
	}
}

func TestNewEventService_ConfigPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := dir + "deployment.yaml"
	dsn := dir + "/events.db"

	yaml := "engines:\n  primary:\n    driver: sqlite\n    dsn: " + dsn + `
instances:
  - role: source-of-truth
    engine: primary
  - role: projections
    engine: primary
`
	writeErr := writeFile(configPath, yaml)
	if writeErr != nil {
		t.Fatalf("write config: %v", writeErr)
	}

	eventSvc, err := NewEventService(EventConfig{ConfigPath: configPath, DLQ: &DLQConfig{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	_, dbErr := eventSvc.DB()
	if dbErr != nil {
		t.Errorf("expected aux DB for the default DLQ store from config-file DSN, got: %v", dbErr)
	}
}

func TestNewEventService_DLQDefaultRequiresSQLite(t *testing.T) {
	t.Parallel()

	_, err := NewEventService(EventConfig{
		Driver: "memory",
		DLQ:    &DLQConfig{Threshold: 2},
	})
	if err == nil {
		t.Fatal("expected error for default DLQ store on non-sqlite driver")
	}

	familyErr, ok := errors.AsType[*errorfamily.Error](err)
	if !ok {
		t.Fatalf("expected *errorfamily.Error, got %T", err)
	}

	if familyErr.Code() != "cqrs.dlq_store_required" {
		t.Errorf("expected code cqrs.dlq_store_required, got %q", familyErr.Code())
	}
}

func TestEventService_DB(t *testing.T) {
	t.Parallel()

	// The aux DB exists for the default DLQ store; checkpoints ride the
	// engine-backed default since ADR-0142.
	eventSvc := newTestEventService(t, EventConfig{DLQ: &DLQConfig{}})
	db, err := eventSvc.DB()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = db.PingContext(context.Background())
	if err != nil {
		t.Fatalf("ping failed: %v", err)
	}
}

func TestEventService_HealthAccessors(t *testing.T) {
	t.Parallel()

	eventSvc := newTestEventService(t, EventConfig{})

	err := eventSvc.HealthCheck(context.Background())
	if err != nil {
		t.Errorf("expected healthy sqlite deployment, got: %v", err)
	}

	engines := eventSvc.EngineHealth(context.Background())
	if len(engines) == 0 {
		t.Fatal("expected at least one engine health entry")
	}

	for _, eng := range engines {
		if eng.Name == "" {
			t.Errorf("engine health entry without a name: %+v", eng)
		}

		if eng.Error != nil {
			t.Errorf("engine %s unhealthy: %v", eng.Name, eng.Error)
		}
	}

	report := eventSvc.ScreamReport()
	if report == nil {
		t.Fatal("expected non-nil ScreamReport")
	}
}

func TestNewEventService_LogsScreamWarningsForVolatileSourceOfTruth(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	eventSvc, err := NewEventService(EventConfig{Driver: memoryDriver, Logger: logger})
	if err != nil {
		t.Fatalf("NewEventService: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	out := buf.String()
	if !strings.Contains(out, "volatile-source-of-truth") {
		t.Errorf("expected volatile-source-of-truth warning at construction, got: %q", out)
	}

	if !strings.Contains(out, "acknowledge_warnings") {
		t.Errorf("expected the acknowledgment escape hatch in the warning, got: %q", out)
	}
}

func TestNewEventService_CleanDeploymentLogsNoScream(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	eventSvc, err := NewEventService(EventConfig{DSN: t.TempDir() + "/test.db", Logger: logger})
	if err != nil {
		t.Fatalf("NewEventService: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	if out := buf.String(); out != "" {
		t.Errorf("expected no safety findings for a clean sqlite deployment, got: %q", out)
	}
}

func TestEventService_Shutdown_Idempotent(t *testing.T) {
	t.Parallel()

	eventSvc, err := NewEventService(EventConfig{
		DSN: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = eventSvc.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("first shutdown: %v", err)
	}

	err = eventSvc.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("second shutdown: %v", err)
	}
}

func TestCloseOnConstructionFailure_NilAuxReturnsErrUntouched(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("primary construction failure")

	result := closeOnConstructionFailure(nil, sentinel)

	if !errors.Is(result, sentinel) {
		t.Errorf("expected sentinel error in result, got: %v", result)
	}
}

func TestCloseOnConstructionFailure_CloseFailureJoinsErrors(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("primary construction failure")

	result := closeOnConstructionFailure(failingCloser{}, sentinel)

	// The primary error must be present in the joined result.
	if !errors.Is(result, sentinel) {
		t.Errorf("expected sentinel error in joined result, got: %v", result)
	}

	// The close error must also be present — both errors surface.
	if !errors.Is(result, errCloseFailed) {
		t.Errorf("expected close error in joined result, got: %v", result)
	}
}

var errCloseFailed = errors.New("simulated close failure")

type failingCloser struct{}

func (failingCloser) Close() error { return errCloseFailed }
