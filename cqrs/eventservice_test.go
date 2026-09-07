package cqrs

import (
	"context"
	"errors"
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

	eventSvc, err := NewEventService(EventConfig{
		DSN: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

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

func TestNewEventService_DeprecatedSQLitePathAlias(t *testing.T) {
	t.Parallel()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	if _, err := eventSvc.DB(); err != nil {
		t.Errorf("expected aux DB via deprecated SQLitePath alias, got: %v", err)
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
	if err := writeFile(configPath, yaml); err != nil {
		t.Fatalf("write config: %v", err)
	}

	eventSvc, err := NewEventService(EventConfig{ConfigPath: configPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	if _, err := eventSvc.DB(); err != nil {
		t.Errorf("expected aux DB from config-file DSN, got: %v", err)
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

	eventSvc, err := NewEventService(EventConfig{
		DSN: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	db, err := eventSvc.DB()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = db.PingContext(context.Background())
	if err != nil {
		t.Fatalf("ping failed: %v", err)
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
