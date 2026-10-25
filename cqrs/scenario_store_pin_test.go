package cqrs

// The bare-ID disk-truth pin (go-cqrs-lite wave I, 2026-10-17): stream IDs
// must persist BARE. The "StreamMarker:" display form leaking into stored
// rows was the class that wave killed across every write path; this pin
// scans EVERY column of EVERY table of a real file-backed SQLite store
// driven through THIS facade and the systemscenario harness — if any layer
// reintroduces the display form (WAL envelope, payload, metadata, entity
// key), the consumer's own suite goes red before any release rides it.

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/systemscenario/v4"
)

func TestScenarioPilot_StoreRowsCarryBareStreamIDs(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "bare-id-pin.db")

	eventSvc := newFacadeServiceCfg(t, EventConfig{DSN: dbPath})

	sc := systemscenario.Adopt(t, context.Background(), eventSvc.System())

	streamID := id.NewStreamID()

	sc.Given().Command(newFacadeCommand(t, streamID)).
		When(newFacadeCommand(t, streamID)).
		Then("facade.bumped")

	if err := eventSvc.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	defer func() { _ = db.Close() }()

	tables, err := queryStringColumn(t, db,
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		t.Fatalf("list tables: %v", err)
	}

	if len(tables) == 0 {
		t.Fatal("no tables in the store — the pin scanned nothing")
	}

	const marker = "StreamMarker:"

	for _, table := range tables {
		scanTableForDisplayForm(t, db, table, marker)
	}

	// The sanity half: the journal must actually carry the stream's events
	// keyed by the BARE id, or the scan above proved nothing.
	var journalTables []string

	for _, table := range tables {
		if strings.Contains(table, "journal") || strings.Contains(table, "stream_log") {
			journalTables = append(journalTables, table)
		}
	}

	if len(journalTables) == 0 {
		t.Fatalf("no journal/stream-log table found among %v — pin cannot prove coverage", tables)
	}
}

func queryStringColumn(t *testing.T, db *sql.DB, query string) ([]string, error) {
	t.Helper()

	rows, err := db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err //nolint:wrapcheck // caller wraps
	}

	defer func() { _ = rows.Close() }()

	var out []string

	for rows.Next() {
		var name string

		if err := rows.Scan(&name); err != nil {
			return nil, err //nolint:wrapcheck // caller wraps
		}

		out = append(out, name)
	}

	return out, rows.Err() //nolint:wrapcheck // caller wraps
}

// quoteSQLiteIdent wraps an identifier in double quotes for the scan query.
func quoteSQLiteIdent(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}

// scanTableForDisplayForm fails the test when any TEXT column of the table
// carries the display-form marker.
func scanTableForDisplayForm(t *testing.T, db *sql.DB, table, marker string) {
	t.Helper()

	rows, err := db.QueryContext(context.Background(), "SELECT * FROM "+quoteSQLiteIdent(table))
	//nolint:gosec // identifier comes from sqlite_master, quoted + escaped
	if err != nil {
		t.Fatalf("scan %s: %v", table, err)
	}

	defer func() { _ = rows.Close() }()

	cols, err := rows.Columns()
	if err != nil {
		t.Fatalf("columns of %s: %v", table, err)
	}

	for rows.Next() {
		raws := make([]any, len(cols)) //nolint:makezero // pre-sized scan targets, populated via ptrs
		ptrs := make([]any, len(cols)) //nolint:makezero // pointers into the pre-sized targets

		for i := range raws {
			ptrs[i] = &raws[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			t.Fatalf("scan row in %s: %v", table, err)
		}

		for i, raw := range raws {
			text, ok := cellText(raw)
			if !ok {
				continue
			}

			if strings.Contains(text, marker) {
				t.Fatalf(
					"%s.%s carries the %q display form — stream IDs must persist bare "+
						"(a write path reintroduced the branded String() form)",
					table, cols[i], marker,
				)
			}
		}
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("iterate %s: %v", table, err)
	}
}

// cellText renders a scanned SQLite cell as text when it is text-shaped.
func cellText(raw any) (string, bool) {
	switch v := raw.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	default:
		return "", false
	}
}
