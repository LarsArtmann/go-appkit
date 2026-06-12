package appkit

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

var allowedPRAGMAs = map[string]bool{
	"journal_mode":       true,
	"busy_timeout":       true,
	"foreign_keys":       true,
	"synchronous":        true,
	"cache_size":         true,
	"temp_store":         true,
	"mmap_size":          true,
	"journal_size_limit": true,
	"wal_autocheckpoint": true,
}

// SQLiteConfig controls the SQLite connection opened by OpenSQLite.
type SQLiteConfig struct {
	Path            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	PRAGMAs         map[string]string
}

// DefaultSQLitePRAGMAs returns a conservative set of SQLite pragmas.
func DefaultSQLitePRAGMAs() map[string]string {
	return map[string]string{
		"journal_mode": "WAL",
		"busy_timeout": "5000",
		"foreign_keys": "ON",
	}
}

// OpenSQLite opens a SQLite connection with sensible defaults.
func OpenSQLite(cfg SQLiteConfig) (*sql.DB, error) {
	if cfg.Path == "" {
		return nil, errors.New("sqlite path is required")
	}

	db, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", cfg.Path, err)
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}

	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	pragmas := cfg.PRAGMAs
	if pragmas == nil {
		pragmas = DefaultSQLitePRAGMAs()
	}

	for key, value := range pragmas {
		if !allowedPRAGMAs[key] {
			_ = db.Close()

			return nil, fmt.Errorf("unsupported PRAGMA %q: not in allowlist", key)
		}

		if _, err := db.Exec(fmt.Sprintf("PRAGMA %s = %s;", key, value)); err != nil {
			_ = db.Close()

			return nil, fmt.Errorf("set PRAGMA %s: %w", key, err)
		}
	}

	return db, nil
}
