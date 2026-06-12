package appkit

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenSQLite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	database, err := OpenSQLite(context.Background(), SQLiteConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = database.Close() }()

	pingErr := database.PingContext(context.Background())
	if pingErr != nil {
		t.Fatalf("ping failed: %v", pingErr)
	}
}

func TestOpenSQLite_EmptyPath(t *testing.T) {
	t.Parallel()

	_, err := OpenSQLite(context.Background(), SQLiteConfig{Path: ""})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestOpenSQLite_DisallowedPRAGMA(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	_, err := OpenSQLite(context.Background(), SQLiteConfig{
		Path:    dbPath,
		PRAGMAs: map[string]string{"drop_table_users": "1"},
	})
	if err == nil {
		t.Fatal("expected error for disallowed PRAGMA key")
	}
}

func TestOpenSQLite_CustomAllowedPRAGMAs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	database, err := OpenSQLite(context.Background(), SQLiteConfig{
		Path: dbPath,
		PRAGMAs: map[string]string{
			"journal_mode": "WAL",
			"cache_size":   "10000",
		},
	})
	if err != nil {
		t.Fatalf("OpenSQLite with custom pragmas failed: %v", err)
	}
	defer func() { _ = database.Close() }()

	pingErr := database.PingContext(context.Background())
	if pingErr != nil {
		t.Fatalf("ping failed: %v", pingErr)
	}
}

func TestOpenSQLite_PoolSettings(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	database, err := OpenSQLite(context.Background(), SQLiteConfig{
		Path:            dbPath,
		MaxOpenConns:    5,
		MaxIdleConns:    3,
		ConnMaxLifetime: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("OpenSQLite with pool settings failed: %v", err)
	}
	defer func() { _ = database.Close() }()

	pingErr := database.PingContext(context.Background())
	if pingErr != nil {
		t.Fatalf("ping failed: %v", pingErr)
	}
}

func TestOpenSQLite_OpenError(t *testing.T) {
	t.Parallel()

	_, err := OpenSQLite(context.Background(), SQLiteConfig{Path: "/nonexistent/dir/test.db"})
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}
