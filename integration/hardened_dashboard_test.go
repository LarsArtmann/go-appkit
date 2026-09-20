package integration_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit"
	appkithealth "github.com/larsartmann/go-appkit/health"
	"github.com/larsartmann/go-appkit/security"
	health "github.com/larsartmann/go-health"
)

// TestHardenedDashboardBehindCSP composes the full hardened dashboard
// posture from the review plan (T21): the health module's
// DashboardHardenedPreset (basePath + per-request nonce extraction) driven
// by the security module's nonce machinery, with the Content-Security-Policy
// itself built OUTSIDE the health module and injected through an OUTER
// middleware — the documented boundary (rate limiting and security headers
// belong in front of the mux; the health module stays core-free).
//
// Asserted: the dashboard answers 200 behind the strict CSP middleware, the
// Content-Security-Policy header is present and carries THIS request's nonce
// in script-src (nonce minted per request, no reuse), and the probe routes
// the dashboard registered still answer.
func TestHardenedDashboardBehindCSP(t *testing.T) {
	t.Parallel()

	probe := health.NewWithHealthCheck(func(context.Context) map[string]error {
		return map[string]error{"database": nil}
	}, health.WithCriticalServices("database"))

	// Bridge the security module's context-based nonce to the preset's
	// request-based extractor contract.
	nonceFn := func(r *http.Request) string { return security.NonceFromContext(r.Context()) }
	mounted, err := appkithealth.New(probe, appkithealth.WithDashboard(
		appkithealth.DashboardHardenedPreset("/health", nonceFn)...,
	))
	if err != nil {
		t.Fatalf("appkithealth.New: %v", err)
	}

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = freeAddr(t)
	cfg.DrainDelay = time.Millisecond
	cfg.OuterMiddlewares = []func(http.Handler) http.Handler{hardenedCSP(t)}

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	mounted.RegisterRoutes(svc.Mux)

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start service: %v", err)
	}
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownErr := svc.Shutdown(shutdownCtx)
		if shutdownErr != nil {
			t.Errorf("shutdown: %v", shutdownErr)
		}
		select {
		case serveErr := <-errCh:
			if serveErr != nil {
				t.Errorf("server returned error: %v", serveErr)
			}
		case <-time.After(2 * time.Second):
			t.Error("server did not stop after shutdown")
		}
	})

	started := time.Now().Add(2 * time.Second)
	for !svc.Running() {
		if time.Now().After(started) {
			t.Fatal("service did not start within timeout")
		}

		time.Sleep(time.Millisecond)
	}

	base := "http://" + svc.Addr().String()

	// The dashboard endpoint must answer under the strict policy: HTML for
	// browser Accept, the JSON status payload otherwise — both prove the
	// route serves through the CSP middleware.
	csp, body := getWithHeaders(t, base+"/health")
	if !strings.Contains(body, "<html") && !strings.Contains(body, `"status"`) {
		t.Fatalf("dashboard body looks empty: %.80s", body)
	}
	if csp == "" {
		t.Fatal("dashboard response carries no Content-Security-Policy header")
	}

	// The header must carry the nonce minted FOR THIS REQUEST: two requests
	// get two different nonce-… script-src tokens (no mint-once reuse).
	first := nonceOf(t, csp)
	csp2, _ := getWithHeaders(t, base+"/health")
	second := nonceOf(t, csp2)
	if first == "" {
		t.Fatalf("script-src carries no nonce: %s", csp)
	}
	if first == second {
		t.Fatalf("CSP nonce reused across requests: %s", first)
	}

	// The dashboard's registered probe routes still answer behind the
	// middleware.
	reqCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, reqErr := http.NewRequestWithContext(reqCtx, http.MethodGet, base+"/health/readyz", nil)
	if reqErr != nil {
		t.Fatalf("request /health/readyz: %v", reqErr)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /health/readyz: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/health/readyz = %d, want 200", resp.StatusCode)
	}
}

// hardenedCSP mints a per-request nonce, exposes it to the handlers through
// the security module's context contract, and sets the matching strict
// policy header — the middleware a real operator runs OUTSIDE the health
// module, in front of the mux.
func hardenedCSP(t *testing.T) func(http.Handler) http.Handler {
	t.Helper()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nonce, nonceErr := security.GenerateNonce()
			if nonceErr != nil {
				http.Error(w, "nonce generation failed", http.StatusInternalServerError)

				return
			}

			policy := security.BuildCSP(security.CSPConfig{
				Environment: security.Production,
				Nonce:       nonce,
			})
			w.Header().Set("Content-Security-Policy", policy)
			next.ServeHTTP(w, r.WithContext(security.WithNonce(r.Context(), nonce)))
		})
	}
}

func getWithHeaders(t *testing.T, url string) (string, string) {
	t.Helper()

	reqCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("request %s: %v", url, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read %s: %v", url, err)
	}

	return resp.Header.Get("Content-Security-Policy"), string(raw)
}

// nonceOf extracts the first `'nonce-…'` token from a script-src directive.
func nonceOf(t *testing.T, csp string) string {
	t.Helper()

	for directive := range strings.SplitSeq(csp, ";") {
		if !strings.HasPrefix(strings.TrimSpace(directive), "script-src") {
			continue
		}

		for token := range strings.FieldsSeq(directive) {
			if nonce, ok := strings.CutPrefix(token, "'nonce-"); ok {
				nonce = strings.TrimSuffix(nonce, "'")

				return nonce
			}
		}
	}

	return ""
}
