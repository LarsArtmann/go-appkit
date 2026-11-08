// The preset composition E2E (P8): preset.New over PUBLISHED tags — the
// one-call spine plus two code-configured legs (realtime, security-standard),
// booted through the full appkit chain. Pins the consumer-facing contract
// against the released module: the dashboard health surface, the version
// stamp, the security-profile header on plain GETs, one live SSE event, and
// a graceful stop with the stream having been drained. Bodies are drained
// before close everywhere (the 2026-11-06 house rule: close-without-drain is
// an RST generator and a shutdown-flake factory).

package integration_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/preset"
	"github.com/larsartmann/go-appkit/realtime"
	"github.com/larsartmann/go-appkit/testkit"
	"github.com/larsartmann/go-sse"
	"github.com/larsartmann/go-sse/ssetest"
)

// presetGet GETs url through a deadline-bounded client, DRAINS the body, and
// returns status, headers, and body text — nothing leaks, nothing lingers.
func presetGet(t *testing.T, url string) (int, http.Header, string) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("build GET %s: %v", url, err)
	}

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	var body strings.Builder

	if _, err := io.Copy(&body, resp.Body); err != nil {
		t.Fatalf("GET %s: drain body: %v", url, err)
	}

	return resp.StatusCode, resp.Header, body.String()
}

// TestPresetE2E_SpineLegsAndGracefulStop exercises the published preset
// module end to end: zero-app-code spine (dashboard health surface + version
// stamp), the realtime leg streaming one event, the security-standard header
// observable on a plain probe, and a graceful shutdown.
func TestPresetE2E_SpineLegsAndGracefulStop(t *testing.T) {
	t.Parallel()

	var hub *realtime.Hub

	// The E2E exercises the default spine plus two legs; zero values are the point.
	cfg := preset.Config{
		Addr:     testkit.FreeAddr(t),
		Version:  "preset-e2e-0.1.0",
		Timeouts: preset.TimeoutsConfig{DrainDelay: appkit.NoDrainDelay},
		Security: &preset.SecurityConfig{
			Enabled: true,
			Profile: preset.SecurityStandard,
		},
		Realtime: &preset.RealtimeConfig{
			Enabled: true,
			Hub:     &hub,
		},
	}

	svc, err := preset.New(cfg)
	if err != nil {
		t.Fatalf("preset.New: %v", err)
	}

	ts := testkit.Serve(t, svc)

	if hub == nil {
		t.Fatal("RealtimeConfig.Hub out-field was not assigned")
	}

	status, header, body := presetGet(t, ts.FullChainURL+"/healthz")
	if status != http.StatusOK {
		t.Errorf("/healthz = %d, want 200", status)
	}

	if header.Get("Strict-Transport-Security") == "" {
		t.Error("security-standard leg: no Strict-Transport-Security on /healthz")
	}

	if strings.Contains(body, "<") {
		t.Errorf("/healthz body %q is not the plain probe payload", body)
	}

	status, _, body = presetGet(t, ts.FullChainURL+"/version")
	if status != http.StatusOK || !strings.Contains(body, "preset-e2e-0.1.0") {
		t.Errorf("/version = %d body %q, want 200 carrying the version stamp", status, body)
	}

	status, header, _ = presetGet(t, ts.FullChainURL+"/health")
	if status != http.StatusOK {
		t.Errorf("dashboard /health = %d, want 200", status)
	}

	if !strings.Contains(header.Get("Content-Type"), "text/html") {
		t.Errorf("dashboard Content-Type = %q, want text/html", header.Get("Content-Type"))
	}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.FullChainURL+"/events", nil)
	if err != nil {
		t.Fatalf("build subscribe request: %v", err)
	}

	sub, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("subscribe /events: %v", err)
	}
	defer func() { _ = sub.Body.Close() }()

	if sub.StatusCode != http.StatusOK {
		t.Fatalf("/events = %d, want 200 streaming", sub.StatusCode)
	}

	if ct := sub.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("/events Content-Type = %q, want text/event-stream", ct)
	}

	deadline := time.Now().Add(5 * time.Second)
	for hub.SubscriberCount() != 1 {
		if time.Now().After(deadline) {
			t.Fatal("handler never subscribed to the hub")
		}

		time.Sleep(5 * time.Millisecond)
	}

	hub.Broadcast(sse.Event{Event: "e2e", Data: `{"hello":"published-preset"}`})

	live := ssetest.MustReadNEvents(t, sub.Body, 1)[0]
	ssetest.RequireEventType(t, live, "e2e")

	// Hub drain FIRST: it ends the stream server-side, so the client sees a
	// clean EOF and the deferred body close sends no RST (the house ordering
	// the realtime module documents for graceful teardown).
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := hub.Shutdown(ctx); err != nil {
		t.Errorf("hub shutdown: %v", err)
	}

	if err := ts.Shutdown(ctx); err != nil {
		t.Errorf("graceful shutdown: %v", err)
	}
}
