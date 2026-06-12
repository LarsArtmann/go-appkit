package appkit

import (
	"path/filepath"
	"testing"
)

func TestOpenSQLite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := OpenSQLite(SQLiteConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		t.Fatalf("ping failed: %v", err)
	}
}

func TestOpenSQLite_EmptyPath(t *testing.T) {
	t.Parallel()

	_, err := OpenSQLite(SQLiteConfig{Path: ""})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}
