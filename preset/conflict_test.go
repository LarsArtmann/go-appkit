package preset_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/larsartmann/go-appkit/preset"
	"github.com/larsartmann/go-appkit/realtime"
)

// The P7 conflict matrix (design note §5): every switch crossed with
// nil/absent, nil/env-on, nil/env-off, code-on+env-off, code-off+env-on,
// and nil/invalid. Boot-level rows with the cheapest deterministic
// observable per leg; deeper behavior (spans, 413, streaming) lives in the
// dedicated leg test files. All rows are serial — they mutate env.

type matrixRow struct {
	name     string
	env      map[string]string
	mutate   func(*preset.Config)
	wantErr  bool
	assert   func(t *testing.T, ts string)
	inactive func(t *testing.T, ts string)
}

func runMatrix(t *testing.T, rows []matrixRow) {
	t.Helper()

	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			for key, value := range row.env {
				t.Setenv(key, value)
			}

			cfg := testConfig()
			if row.mutate != nil {
				row.mutate(&cfg)
			}

			svc, err := preset.New(cfg)
			if row.wantErr {
				if err == nil {
					t.Fatal("New succeeded, want boot Rejection")
				}

				return
			}
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			ts := servePreset(t, svc)

			if row.assert != nil {
				row.assert(t, ts)
			}

			if row.inactive != nil {
				row.inactive(t, ts)
			}
		})
	}
}

// pretty404 asserts the errorpages catch-all answers unknown paths.
func pretty404(t *testing.T, ts string) {
	t.Helper()

	_, contentType, body := getNotFound(t, ts, "")

	if !strings.Contains(contentType, "text/html") {
		t.Errorf("unknown path Content-Type = %q, want the pretty HTML 404", contentType)
	}

	if strings.Contains(body, "404 page not found") {
		t.Error("stdlib 404 served — the errorpages leg is inactive")
	}
}

// plain404 asserts NO catch-all is mounted.
func plain404(t *testing.T, ts string) {
	t.Helper()

	_, _, body := getNotFound(t, ts, "")

	if !strings.Contains(body, "404 page not found") {
		t.Errorf("unknown path body %q — the catch-all is still mounted", body)
	}
}

// hstsPresent asserts the standard-profile production posture joined.
func hstsPresent(t *testing.T, ts string) {
	t.Helper()

	resp := getWithHeaders(t, ts+"/healthz")
	defer func() { _ = resp.Body.Close() }()

	if resp.Header.Get("Strict-Transport-Security") == "" {
		t.Error("HSTS missing — the security leg is inactive")
	}
}

// hstsAbsent asserts no profile joined the stack.
func hstsAbsent(t *testing.T, ts string) {
	t.Helper()

	resp := getWithHeaders(t, ts+"/healthz")
	defer func() { _ = resp.Body.Close() }()

	if resp.Header.Get("Strict-Transport-Security") != "" {
		t.Error("HSTS present — the security leg is active")
	}
}

func TestConflictMatrix_ErrorPages(t *testing.T) { //nolint:dupl // six-row grammar is uniform across legs by design (design note §5)
	runMatrix(t, []matrixRow{
		{name: "absent-defaults-on", assert: pretty404},
		{name: "env-on", env: map[string]string{"APPKIT_ERRORPAGES": "on"}, assert: pretty404},
		{name: "env-off", env: map[string]string{"APPKIT_ERRORPAGES": "off"}, assert: plain404},
		{
			name: "code-on-env-off",
			env:  map[string]string{"APPKIT_ERRORPAGES": "off"},
			mutate: func(c *preset.Config) {
				c.ErrorPages = &preset.ErrorPagesConfig{Enabled: true}
			},
			assert: pretty404,
		},
		{
			name: "code-off-env-on",
			env:  map[string]string{"APPKIT_ERRORPAGES": "on"},
			mutate: func(c *preset.Config) {
				c.ErrorPages = &preset.ErrorPagesConfig{Enabled: false}
			},
			assert: plain404,
		},
		{name: "env-invalid", env: map[string]string{"APPKIT_ERRORPAGES": "banana"}, wantErr: true},
	})
}

func TestConflictMatrix_Security(t *testing.T) { //nolint:dupl // six-row grammar is uniform across legs by design (design note §5)
	runMatrix(t, []matrixRow{
		{name: "absent-defaults-none", assert: hstsAbsent},
		{name: "env-none", env: map[string]string{"APPKIT_SECURITY": "none"}, assert: hstsAbsent},
		{name: "env-standard", env: map[string]string{"APPKIT_SECURITY": "standard"}, assert: hstsPresent},
		{
			name: "code-on-env-none",
			env:  map[string]string{"APPKIT_SECURITY": "none"},
			mutate: func(c *preset.Config) {
				c.Security = &preset.SecurityConfig{Enabled: true}
			},
			assert: hstsPresent,
		},
		{
			name: "code-off-env-standard",
			env:  map[string]string{"APPKIT_SECURITY": "standard"},
			mutate: func(c *preset.Config) {
				c.Security = &preset.SecurityConfig{Enabled: false}
			},
			assert: hstsAbsent,
		},
		{name: "env-invalid", env: map[string]string{"APPKIT_SECURITY": "maybe"}, wantErr: true},
	})
}

func TestConflictMatrix_Realtime(t *testing.T) {
	assignHub := func(c *preset.Config, hub **realtime.Hub) {
		c.Realtime = &preset.RealtimeConfig{Enabled: true, Hub: hub}
	}

	eventsMounted := func(t *testing.T, ts string) {
		t.Helper()

		resp := getWithHeaders(t, ts+"/events")
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("/events = %d, want 200 (stream)", resp.StatusCode)
		}
	}

	eventsAbsent := func(t *testing.T, ts string) {
		t.Helper()

		resp := getWithHeaders(t, ts+"/events")
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("/events = %d, want 404 (leg inactive)", resp.StatusCode)
		}
	}

	var codeOnHub, codeOffHub *realtime.Hub

	runMatrix(t, []matrixRow{
		{name: "absent-defaults-off", assert: eventsAbsent},
		{name: "env-on", env: map[string]string{"APPKIT_REALTIME": "on"}, assert: eventsMounted},
		{name: "env-off", env: map[string]string{"APPKIT_REALTIME": "off"}, assert: eventsAbsent},
		{
			name: "code-on-env-off",
			env:  map[string]string{"APPKIT_REALTIME": "off"},
			mutate: func(c *preset.Config) {
				assignHub(c, &codeOnHub)
			},
			assert: eventsMounted,
		},
		{
			name: "code-off-env-on",
			env:  map[string]string{"APPKIT_REALTIME": "on"},
			mutate: func(c *preset.Config) {
				c.Realtime = &preset.RealtimeConfig{Enabled: false, Hub: &codeOffHub}
			},
			assert: eventsAbsent,
		},
		{name: "env-invalid", env: map[string]string{"APPKIT_REALTIME": "banana"}, wantErr: true},
	})

	if codeOnHub == nil {
		t.Error("Hub out-field not assigned on the code-enabled row")
	}

	if codeOffHub != nil {
		t.Error("Hub out-field assigned on the code-disabled row")
	}
}

func TestConflictMatrix_OTEL(t *testing.T) {
	boots := func(*testing.T, string) {}

	runMatrix(t, []matrixRow{
		{name: "absent-defaults-off", assert: boots},
		{name: "env-on", env: map[string]string{"APPKIT_OTEL": "on"}, assert: boots},
		{name: "env-off", env: map[string]string{"APPKIT_OTEL": "off"}, assert: boots},
		{
			name: "code-on-env-off",
			env:  map[string]string{"APPKIT_OTEL": "off"},
			mutate: func(c *preset.Config) {
				c.OTEL = &preset.OTELConfig{Enabled: true}
			},
			assert: boots,
		},
		{
			name: "code-off-env-on",
			env:  map[string]string{"APPKIT_OTEL": "on"},
			mutate: func(c *preset.Config) {
				c.OTEL = &preset.OTELConfig{Enabled: false}
			},
			assert: boots,
		},
		{name: "env-invalid", env: map[string]string{"APPKIT_OTEL": "banana"}, wantErr: true},
	})
}
