package preset_test

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit/preset"
	"github.com/larsartmann/go-appkit/realtime"
	"github.com/larsartmann/go-appkit/testkit"
	"github.com/larsartmann/go-sse"
	"github.com/larsartmann/go-sse/ssetest"
)

// TestRealtime_StreamsEvents: the enabled leg mounts GET /events, hands the
// constructed hub to the consumer via the Hub out-field, and broadcasts reach
// established subscribers. Sync on SubscriberCount before broadcasting (the
// SSE house rule), and the hub drains at graceful shutdown.
func TestRealtime_StreamsEvents(t *testing.T) {
	var hub *realtime.Hub

	cfg := testConfig()
	cfg.Realtime = &preset.RealtimeConfig{
		Enabled: true,
		Hub:     &hub,
	}

	svc, err := preset.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ts := testkit.Serve(t, svc)

	if hub == nil {
		t.Fatal("Hub out-field was not assigned")
	}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.FullChainURL+"/events", nil)
	if err != nil {
		t.Fatalf("build subscribe request: %v", err)
	}

	sub, err := (&http.Client{Timeout: httpTimeout}).Do(req)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
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

	hub.Broadcast(sse.Event{Event: "preset", Data: `{"hello":"events"}`})

	live := ssetest.MustReadNEvents(t, sub.Body, 1)[0]
	ssetest.RequireEventType(t, live, "preset")

	if err := ts.Shutdown(t.Context()); err != nil {
		t.Fatalf("graceful shutdown with an open stream: %v", err)
	}
}

// TestRealtime_OffByDefault: no env, nil leg — GET /events is not mounted
// (the errorpages catch-all answers).
func TestRealtime_OffByDefault(t *testing.T) {
	svc, err := preset.New(testConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts := servePreset(t, svc)

	resp := testkit.HTTPGetContext(t, t.Context(), ts+"/events")
	defer func() { _ = resp.Body.Close() }() // respdrain:ignore SSE stream — Close IS the disconnect

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("/events = %d, want 404 (realtime off by default)", resp.StatusCode)
	}
}

// TestRealtime_CustomPattern: Pattern overrides the default mount point.
func TestRealtime_CustomPattern(t *testing.T) {
	cfg := testConfig()
	cfg.Realtime = &preset.RealtimeConfig{
		Enabled: true,
		Pattern: "GET /live-stream",
	}

	svc, err := preset.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ts := testkit.Serve(t, svc)

	mounted := testkit.HTTPGetContext(t, t.Context(), ts.FullChainURL+"/live-stream")
	defer func() { _ = mounted.Body.Close() }() // respdrain:ignore SSE stream — Close IS the disconnect

	if mounted.StatusCode != http.StatusOK {
		t.Errorf("/live-stream = %d, want 200", mounted.StatusCode)
	}

	// The default pattern is gone.
	absent := testkit.HTTPGetContext(t, t.Context(), ts.FullChainURL+"/events")
	defer func() { _ = absent.Body.Close() }() // respdrain:ignore SSE stream — Close IS the disconnect

	if absent.StatusCode != http.StatusNotFound {
		t.Errorf("/events = %d, want 404 (replaced by the custom pattern)", absent.StatusCode)
	}
}

// TestRealtime_EnvInvalid_Rejects: garbage APPKIT_REALTIME fails boot before
// any listener binds.
func TestRealtime_EnvInvalid_Rejects(t *testing.T) {
	t.Setenv("APPKIT_REALTIME", "banana")

	_, err := preset.New(testConfig())
	if err == nil {
		t.Fatal("New succeeded with APPKIT_REALTIME=banana, want Rejection")
	}
}
