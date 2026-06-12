package appkit

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const (
	pragmaJournalMode       = "journal_mode"
	pragmaBusyTimeout       = "busy_timeout"
	pragmaForeignKeys       = "foreign_keys"
	pragmaSynchronous       = "synchronous"
	pragmaCacheSize         = "cache_size"
	pragmaTempStore         = "temp_store"
	pragmaMmapSize          = "mmap_size"
	pragmaJournalSizeLimit  = "journal_size_limit"
	pragmaWalAutocheckpoint = "wal_autocheckpoint"

	pragWAL = "WAL"
	prag5s  = "5000"
	pragOn  = "ON"
)

func allowedPRAGMAKeys() map[string]struct{} {
	return map[string]struct{}{
		pragmaJournalMode:       {},
		pragmaBusyTimeout:       {},
		pragmaForeignKeys:       {},
		pragmaSynchronous:       {},
		pragmaCacheSize:         {},
		pragmaTempStore:         {},
		pragmaMmapSize:          {},
		pragmaJournalSizeLimit:  {},
		pragmaWalAutocheckpoint: {},
	}
}

var (
	errSQLitePathRequired = errors.New("sqlite path is required")
	errPRAGMAAllowlist    = errors.New("unsupported PRAGMA: not in allowlist")
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
		pragmaJournalMode: pragWAL,
		pragmaBusyTimeout: prag5s,
		pragmaForeignKeys: pragOn,
	}
}

// OpenSQLite opens a SQLite connection with sensible defaults.
func OpenSQLite(ctx context.Context, cfg SQLiteConfig) (*sql.DB, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("%w", errSQLitePathRequired)
	}

	database, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", cfg.Path, err)
	}

	if cfg.MaxOpenConns > 0 {
		database.SetMaxOpenConns(cfg.MaxOpenConns)
	}

	if cfg.MaxIdleConns > 0 {
		database.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	if cfg.ConnMaxLifetime > 0 {
		database.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	pragmas := cfg.PRAGMAs
	if pragmas == nil {
		pragmas = DefaultSQLitePRAGMAs()
	}

	for key, value := range pragmas {
		allowed := allowedPRAGMAKeys()
		if _, ok := allowed[key]; !ok {
			_ = database.Close()

			return nil, fmt.Errorf("%w: %q", errPRAGMAAllowlist, key)
		}

		_, execErr := database.ExecContext(ctx, fmt.Sprintf("PRAGMA %s = %s;", key, value))
		if execErr != nil {
			_ = database.Close()

			return nil, fmt.Errorf("set PRAGMA %s: %w", key, execErr)
		}
	}

	return database, nil
}
