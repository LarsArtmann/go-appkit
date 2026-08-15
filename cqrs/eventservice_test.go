package cqrs

import (
	"context"
	"errors"
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
)

func TestNewEventService_EmptyPath(t *testing.T) {
	t.Parallel()

	_, err := NewEventService(EventConfig{})
	if err == nil {
		t.Fatal("expected error for empty SQLitePath")
	}
}

func TestNewEventService_ValidPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: dir + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	if eventSvc.Bundle() == nil {
		t.Fatal("expected non-nil Bundle")
	}

	if eventSvc.Host() == nil {
		t.Fatal("expected non-nil Host")
	}

	if eventSvc.Bundle().EventSink == nil {
		t.Error("expected non-nil EventSink")
	}

	if eventSvc.Bundle().EventSource == nil {
		t.Error("expected non-nil EventSource")
	}

	if eventSvc.Bundle().Publisher == nil {
		t.Error("expected non-nil Publisher")
	}

	if eventSvc.Bundle().Subscriber == nil {
		t.Error("expected non-nil Subscriber")
	}

	if eventSvc.Bundle().CheckpointStore == nil {
		t.Error("expected non-nil CheckpointStore")
	}
}

func TestEventService_DB(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: dir + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	db, err := eventSvc.DB()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("ping failed: %v", err)
	}
}

func TestEventService_Shutdown_Idempotent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: dir + "/test.db",
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

func TestAsSQLDB_RejectsNonSQLDB(t *testing.T) {
	t.Parallel()

	db, err := asSQLDB("not a database")
	if err == nil {
		t.Fatal("expected error for non *sql.DB value")
	}

	if db != nil {
		t.Errorf("expected nil *sql.DB, got %v", db)
	}

	var familyErr *errorfamily.Error
	if !errors.As(err, &familyErr) {
		t.Fatalf("expected *errorfamily.Error, got %T", err)
	}

	if familyErr.Family() != errorfamily.Rejection {
		t.Errorf("expected family %q, got %q", errorfamily.Rejection, familyErr.Family())
	}

	if familyErr.Code() != "cqrs.db_not_sql" {
		t.Errorf("expected code cqrs.db_not_sql, got %q", familyErr.Code())
	}

	if got := errorfamily.HTTPStatus(err); got != 400 {
		t.Errorf("expected HTTP status 400 for Rejection, got %d", got)
	}
}

func TestAsSQLDB_AcceptsSQLDB(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	eventSvc, err := NewEventService(EventConfig{
		SQLitePath: dir + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = eventSvc.Shutdown(context.Background()) }()

	db, err := asSQLDB(eventSvc.Bundle().Database())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if db == nil {
		t.Fatal("expected non-nil *sql.DB")
	}
}
