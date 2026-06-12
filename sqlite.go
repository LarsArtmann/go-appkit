package appkit

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

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
		return nil, fmt.Errorf("sqlite path is required")
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
		if _, err := db.Exec(fmt.Sprintf("PRAGMA %s = %s;", key, value)); err != nil {
			_ = db.Close()

			return nil, fmt.Errorf("set PRAGMA %s: %w", key, err)
		}
	}

	return db, nil
}
