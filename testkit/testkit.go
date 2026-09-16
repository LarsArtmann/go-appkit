// Package testkit provides the full-chain test harness for appkit services.
//
// The trap it encodes as API: a test server built with
// httptest.NewServer(svc.Mux) BYPASSES the entire middleware chain —
// recovery, request IDs, logging, timeout, security headers — and skips
// every route registered by the framework. Three of CV's production bugs
// were only visible through the full chain. [Serve] makes the full chain
// the default: it starts the REAL service on its own listener, so tests
// exercise exactly what production serves.
//
// Teardown order (via t.Cleanup): wait for shutdown, then assert the
// goroutine count returned to baseline — a server that leaks ~25 goroutines
// makes every subsequent test in the package load-flaky.
package testkit

import (
	"context"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit"
)

const (
	// startTimeout bounds how long Serve waits for the listener to come up.
	startTimeout = 5 * time.Second

	// stopTimeout bounds the graceful shutdown in cleanup.
	stopTimeout = 5 * time.Second

	// goroutineTolerance is the allowed overshoot above baseline after
	// shutdown (runtime internals jitter by a couple).
	goroutineTolerance = 5

	// leakPollInterval is the goroutine-count re-check cadence.
	leakPollInterval = 10 * time.Millisecond
)

// TestServer is a running service under test.
type TestServer struct {
	service *appkit.Service

	// FullChainURL is the base URL through the REAL listener: the complete
	// middleware chain plus everything registered on the mux. This is the
	// URL tests should hit.
	FullChainURL string

	goroutineBaseline int
}

// Serve starts svc and registers teardown with t.Cleanup: a graceful
// shutdown, then a goroutine-baseline assertion (tolerance 5) so a leaked
// goroutine fails the test instead of flaking the suite.
func Serve(tb testing.TB, svc *appkit.Service) *TestServer {
	tb.Helper()

	baseline := runtime.NumGoroutine()

	errCh, err := svc.Start()
	if err != nil {
		tb.Fatalf("testkit.Serve: start: %v", err)
	}

	deadline := time.Now().Add(startTimeout)
	for !svc.Running() {
		if time.Now().After(deadline) {
			tb.Fatal("testkit.Serve: service did not start within 5s")
		}

		time.Sleep(time.Millisecond)
	}

	server := &TestServer{
		service:           svc,
		FullChainURL:      "http://" + svc.Addr().String(),
		goroutineBaseline: baseline,
	}

	tb.Cleanup(func() {
		ctx, cancel := contextWithTimeout(stopTimeout)
		defer cancel()

		if err := svc.Shutdown(ctx); err != nil {
			tb.Errorf("testkit.Serve: shutdown: %v", err)
		}

		select {
		case err := <-errCh:
			if err != nil {
				tb.Errorf("testkit.Serve: server error: %v", err)
			}
		case <-time.After(2 * time.Second):
			tb.Error("testkit.Serve: server did not stop after shutdown")
		}

		// Goroutine-baseline assert: a leaked goroutine (an evict loop, a
		// stuck SSE subscriber) fails THIS test instead of flaking the next.
		deadline := time.Now().Add(2 * time.Second)
		for runtime.NumGoroutine() > server.goroutineBaseline+goroutineTolerance {
			if time.Now().After(deadline) {
				tb.Errorf("testkit.Serve: goroutine leak: %d goroutines after shutdown, baseline %d",
					runtime.NumGoroutine(), server.goroutineBaseline)

				break
			}

			time.Sleep(leakPollInterval)
		}
	})

	return server
}

// Mux exposes the service's mux for registration AFTER Serve (routes must
// be registered before Serve in normal use — the mux serves live once the
// service starts).
func (ts *TestServer) Mux() *http.ServeMux { return ts.service.Mux }

// contextWithTimeout is a tiny indirection keeping the cleanup block terse.
func contextWithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}
