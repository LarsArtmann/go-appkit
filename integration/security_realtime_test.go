// The security × realtime composition: a keyed rate limiter in front of the
// live SSE endpoint. Pins the 429-aborts-chain contract against a real
// stream through the appkit default stack: the first subscriber connects and
// keeps receiving broadcasts, a reconnect from the same key gets 429 with
// Retry-After and NO SSE stream — the limiter aborts before the realtime
// handler can ever respond, so the 429 cannot be overwritten.
package integration_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/realtime"
	"github.com/larsartmann/go-appkit/security"
	"github.com/larsartmann/go-appkit/testkit"
	"github.com/larsartmann/go-sse"
	"github.com/larsartmann/go-sse/ssetest"
)

// newRateLimitedSSEService starts an appkit Service (NoTimeout for the
// stream, 1ms drain per the integration house rules) with the hub mounted
// BEHIND a one-token-per-minute limiter keyed by remote host. Every request
// from the test client shares the same key, so the second connect exhausts
// the bucket. Teardown drains the hub BEFORE the service (browsers reconnect
// to another instance while the listener is still up).
func newRateLimitedSSEService(t *testing.T) (*testkit.TestServer, *realtime.Hub) {
	t.Helper()

	hub := realtime.NewHub()

	svc, err := appkit.ServiceConfig{
		Addr:         freeAddr(t),
		ReadTimeout:  appkit.NoTimeout,
		WriteTimeout: appkit.NoTimeout,
		DrainDelay:   time.Millisecond,
	}.New()
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	limited := security.RateLimit(security.RateLimitConfig{
		Limit:        1,
		Burst:        1,
		Window:       time.Minute,
		MaxKeys:      16,
		KeyExtractor: nil,
	})
	svc.Mux.Handle("GET /events", limited(realtime.Handler(hub)))

	ts := testkit.Serve(t, svc)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := hub.Shutdown(ctx); err != nil {
			t.Errorf("hub shutdown: %v", err)
		}

		if err := ts.Shutdown(ctx); err != nil {
			t.Errorf("service shutdown: %v", err)
		}
	})

	return ts, hub
}

func TestRateLimitInFrontOfSSE(t *testing.T) {
	t.Parallel()

	ts, hub := newRateLimitedSSEService(t)

	// First connection from the shared localhost key consumes the single
	// token and establishes a live stream.
	sub, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.FullChainURL+"/events", nil)
	if err != nil {
		t.Fatalf("build subscribe request: %v", err)
	}

	subResp, err := http.DefaultClient.Do(sub)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer subResp.Body.Close()

	if subResp.StatusCode != http.StatusOK {
		t.Fatalf("first connect = %d, want 200 streaming", subResp.StatusCode)
	}

	if ct := subResp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Errorf("first connect Content-Type = %q, want text/event-stream", ct)
	}

	// A live broadcast reaches the established subscriber: the limiter
	// meters CONNECTIONS, not events — it must never throttle a stream
	// that is already open.
	hub.Broadcast(sse.Event{Event: "live", Data: "{}", ID: sse.NewEventID("sse-rate-1")})

	live := ssetest.MustReadNEvents(t, subResp.Body, 1)[0]
	ssetest.RequireEventID(t, live, "sse-rate-1")

	// Reconnect from the same key: 429 with Retry-After, and the chain
	// ABORTED — the realtime handler never runs, so the response carries
	// no SSE stream and the 429 cannot be overwritten.
	retry, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.FullChainURL+"/events", nil)
	if err != nil {
		t.Fatalf("build reconnect request: %v", err)
	}

	retryResp, err := http.DefaultClient.Do(retry)
	if err != nil {
		t.Fatalf("reconnect: %v", err)
	}
	defer retryResp.Body.Close()

	if retryResp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("reconnect = %d, want 429", retryResp.StatusCode)
	}

	if retryResp.Header.Get("Retry-After") == "" {
		t.Error("429 missing Retry-After header")
	}

	if ct := retryResp.Header.Get("Content-Type"); strings.HasPrefix(ct, "text/event-stream") {
		t.Error("429 carries an SSE stream — the limiter did not abort the chain before realtime.Handler")
	}
}
