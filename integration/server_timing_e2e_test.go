// The core v0.10.0 Server-Timing leg through a live appkit Service against
// PUBLISHED tags (the R4b ride-along on P8's re-pin): the default stack
// stamps one W3C Server-Timing entry (total;dur=<ms>) on every response,
// and the `ServerTiming` knob (RegisterHealth *bool semantics: nil or true
// keeps it, &false omits the middleware entirely) removes it. Wire contract
// pinned at the composition level — what a real consumer's DevTools/APM sees:
//
//   - default config → `Server-Timing: total;dur=<number>` on plain GETs
//   - ServerTiming: &false → no Server-Timing header at all
//   - the header is additive: statuses and bodies are unaffected
//
// Bodies are drained before close everywhere (the 2026-11-06 house rule).

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
)

func newTimingService(t *testing.T, knob *bool) *testkit.TestServer {
	t.Helper()

	svc, err := appkit.NewService(appkit.ServiceConfig{
		Addr:         testkit.FreeAddr(t),
		DrainDelay:   time.Millisecond,
		ServerTiming: knob,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	svc.Mux.HandleFunc("GET /ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ts := testkit.Serve(t, svc)

	return ts
}

func TestServerTimingDefaultStackHeader(t *testing.T) {
	t.Parallel()

	ts := newTimingService(t, nil)

	resp := testkit.HTTPGetURL(t, ts.FullChainURL+"/ping")
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body) // drain: an unread body is an RST generator
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /ping: status = %d, want 200", resp.StatusCode)
	}

	st := resp.Header.Get("Server-Timing")
	if st == "" {
		t.Fatal("GET /ping: no Server-Timing header on the default stack (total stamp missing)")
	}
	if !strings.Contains(st, "total") {
		t.Errorf("Server-Timing = %q, want it to name the total entry", st)
	}
	if !strings.Contains(st, "dur=") {
		t.Errorf("Server-Timing = %q, want a dur= duration", st)
	}
}

func TestServerTimingKnobOmits(t *testing.T) {
	t.Parallel()

	ts := newTimingService(t, &[]bool{false}[0])

	resp := testkit.HTTPGetURL(t, ts.FullChainURL+"/ping")
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body) // drain: an unread body is an RST generator
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /ping: status = %d, want 200", resp.StatusCode)
	}

	if st := resp.Header.Get("Server-Timing"); st != "" {
		t.Errorf("GET /ping with ServerTiming=&false: Server-Timing = %q, want absent", st)
	}
}

// TestServerTimingThroughShutdown closes the loop with the teardown: the
// header stamps at response commit (SSE-safe per the middleware contract),
// and a service that served stamped responses still shuts down cleanly —
// the explicit Shutdown assertion keeps the wire contract tied to graceful
// stop.
func TestServerTimingThroughShutdown(t *testing.T) {
	t.Parallel()

	ts := newTimingService(t, nil)

	resp := testkit.HTTPGetURL(t, ts.FullChainURL+"/ping")
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	http.DefaultClient.CloseIdleConnections()

	if shutdownErr := ts.Shutdown(ctx); shutdownErr != nil {
		t.Errorf("shutdown after stamped responses: %v", shutdownErr)
	}
}
