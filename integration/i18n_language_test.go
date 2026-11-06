// The httputil Language() negotiation leg through the appkit default
// middleware stack (Recovery → RequestID → Logging → Timeout →
// SecurityHeaders). Pins the composition the i18n-fleet decision memo left
// demand-gated (trigger FIRED 2026-10-25: httputil v1.6.0 published
// Language(); the fleet pins replace-free v1.7.0):
//
//   - Accept-Language negotiation works under the full default stack —
//     primary-subtag fallback ("da-DK" serves "da") and q-weight ordering
//     both land on the configured supported tags.
//   - Unsupported candidates negotiate to the default tag ("fr" → "en").
//   - Missing header negotiates to the default tag.
//   - Content-Language and Vary: Accept-Language ride every negotiated
//     response THROUGH the default stack (no middleware strips them).
//   - LanguageFromContext is populated for downstream handlers.

package integration_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/testkit"
	"github.com/larsartmann/httputil"
)

func newLanguageService(t *testing.T) *testkit.TestServer {
	t.Helper()

	svc, err := appkit.NewService(appkit.ServiceConfig{
		Addr:       freeAddr(t),
		DrainDelay: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	langCfg := httputil.DefaultLanguageConfig()
	langCfg.SupportedTags = []string{"en", "da"}

	greet := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("lang=" + httputil.LanguageFromContext(r.Context())))
	})
	svc.Mux.Handle("GET /greet", httputil.Language(langCfg)(greet))

	ts := testkit.Serve(t, svc)

	t.Cleanup(func() {
		// Close client pools BEFORE the server waits: a spare conn dialed by
		// a shared Transport under parallel-burst contention leaves a
		// StateNew server conn, and net/http's Shutdown refuses to close a
		// StateNew conn until it is 5s old (issue 22682 grace) — the exact
		// budget of this ctx (the 2026-11-06 moving shutdown-flake class).
		http.DefaultClient.CloseIdleConnections()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if shutdownErr := ts.Shutdown(ctx); shutdownErr != nil {
			t.Errorf("shutdown: %v", shutdownErr)
		}
	})

	return ts
}

// negotiate GETs /greet with the given Accept-Language, closes the response
// body, and returns the response headers plus the negotiated language tag
// the handler echoed back.
// langProbe is shared across all negotiate calls: one connection pool.
// Abandoned per-call transports hold idle keep-alive conns open against the
// shutdown window (the 2026-11-06 listener_close hang class).
var langProbe = &http.Client{Timeout: 10 * time.Second}

func negotiate(t *testing.T, base, acceptLanguage string) (http.Header, string) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, base+"/greet", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if acceptLanguage != "" {
		req.Header.Set("Accept-Language", acceptLanguage)
	}

	resp, err := langProbe.Do(req)
	if err != nil {
		t.Fatalf("GET /greet: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /greet (Accept-Language %q): status = %d, want 200", acceptLanguage, resp.StatusCode)
	}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		t.Fatalf("read body: %v", readErr)
	}

	return resp.Header, strings.TrimPrefix(strings.TrimSpace(string(body)), "lang=")
}

func TestLanguageNegotiationThroughDefaultStack(t *testing.T) {
	t.Parallel()

	ts := newLanguageService(t)
	base := ts.FullChainURL

	t.Run("primary-subtag fallback and q-weight ordering", func(t *testing.T) {
		t.Parallel()

		header, lang := negotiate(t, base, "da-DK,da;q=0.9,en;q=0.4")
		if lang != "da" {
			t.Errorf("negotiated = %q, want da (primary-subtag fallback)", lang)
		}
		if cl := header.Get("Content-Language"); cl != "da" {
			t.Errorf("Content-Language = %q, want da", cl)
		}
		if vary := header.Get("Vary"); !strings.Contains(vary, "Accept-Language") {
			t.Errorf("Vary = %q, want it to name Accept-Language", vary)
		}

		header2, lang2 := negotiate(t, base, "en;q=0.2,da;q=0.9")
		if lang2 != "da" {
			t.Errorf("q-weight negotiated = %q, want da (0.9 beats 0.2)", lang2)
		}
		if cl := header2.Get("Content-Language"); cl != "da" {
			t.Errorf("q-weight Content-Language = %q, want da", cl)
		}
	})

	t.Run("unsupported candidates fall to the default tag", func(t *testing.T) {
		t.Parallel()

		_, lang := negotiate(t, base, "fr-FR,fr;q=0.9")
		if lang != "en" {
			t.Errorf("negotiated = %q, want en (unsupported fr falls to default)", lang)
		}
	})

	t.Run("missing header negotiates the default tag", func(t *testing.T) {
		t.Parallel()

		_, lang := negotiate(t, base, "")
		if lang != "en" {
			t.Errorf("negotiated = %q, want en (no Accept-Language)", lang)
		}
	})
}
