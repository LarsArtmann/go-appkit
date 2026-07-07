package cqrs

import (
	"context"
	"testing"
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
