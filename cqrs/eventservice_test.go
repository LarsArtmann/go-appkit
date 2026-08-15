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

	es, err := NewEventService(EventConfig{
		SQLitePath: dir + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = es.Shutdown(context.Background()) }()

	if es.Bundle() == nil {
		t.Fatal("expected non-nil Bundle")
	}

	if es.Host() == nil {
		t.Fatal("expected non-nil Host")
	}

	if es.Bundle().EventSink == nil {
		t.Error("expected non-nil EventSink")
	}

	if es.Bundle().EventSource == nil {
		t.Error("expected non-nil EventSource")
	}

	if es.Bundle().Publisher == nil {
		t.Error("expected non-nil Publisher")
	}

	if es.Bundle().Subscriber == nil {
		t.Error("expected non-nil Subscriber")
	}

	if es.Bundle().CheckpointStore == nil {
		t.Error("expected non-nil CheckpointStore")
	}
}

func TestEventService_DB(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	es, err := NewEventService(EventConfig{
		SQLitePath: dir + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = es.Shutdown(context.Background()) }()

	db, err := es.DB()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("ping failed: %v", err)
	}
}

func TestEventService_Shutdown_Idempotent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	es, err := NewEventService(EventConfig{
		SQLitePath: dir + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := es.Shutdown(context.Background()); err != nil {
		t.Fatalf("first shutdown: %v", err)
	}

	if err := es.Shutdown(context.Background()); err != nil {
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

	es, err := NewEventService(EventConfig{
		SQLitePath: dir + "/test.db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() { _ = es.Shutdown(context.Background()) }()

	db, err := asSQLDB(es.Bundle().Database())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if db == nil {
		t.Fatal("expected non-nil *sql.DB")
	}
}
