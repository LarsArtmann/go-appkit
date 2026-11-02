package testkit

import (
	"context"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit"
)

// waitRunningTimeout bounds WaitForRunning's poll for the listener to come
// up. Generous enough for saturated shared boxes (the load-flake class);
// a service that has not started in 5s has failed for a real reason.
const waitRunningTimeout = 5 * time.Second

// HTTPGetURL issues a context-aware GET against a test server and returns
// the response. Request-construction and transport errors are fatal with
// the URL in the message. The caller owns closing the response body.
func HTTPGetURL(tb testing.TB, url string) *http.Response {
	tb.Helper()

	return HTTPGetContext(tb, context.Background(), url)
}

// HTTPGetContext is HTTPGetURL with a caller-supplied context, for tests
// that need cancellation or deadlines on the request itself.
func HTTPGetContext(tb testing.TB, ctx context.Context, url string) *http.Response {
	tb.Helper()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		tb.Fatalf("testkit: build GET %s: %v", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		tb.Fatalf("testkit: GET %s: %v", url, err)
	}

	return resp
}

// WaitForRunning polls svc.Running until true or the 5s bound; a service
// that never started fails the test.
func WaitForRunning(tb testing.TB, svc *appkit.Service) {
	tb.Helper()

	WaitForRunningWithin(tb, svc, waitRunningTimeout)
}

// WaitForRunningWithin is WaitForRunning with an explicit bound.
func WaitForRunningWithin(tb testing.TB, svc *appkit.Service, timeout time.Duration) {
	tb.Helper()

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if svc.Running() {
			return
		}

		time.Sleep(time.Millisecond)
	}

	tb.Fatalf("testkit: service did not start within %s", timeout)
}

// freeAddr bookkeeping: the issued registry plus the rebind probe. The
// registry closes the cross-test race where two parallel tests draw the
// same ephemeral port before either binds; the rebind probe closes the
// steal-between-close-and-bind race (the 2026-10-24 runner EADDRINUSE
// class). Both checks redraw instead of failing the test later.
//
//nolint:gochecknoglobals // process-lifetime issued-port registry; the dedup is the point
var (
	freeAddrMu     sync.Mutex
	freeAddrIssued = map[string]struct{}{}
)

// FreeAddr reserves a loopback address for a service: an ephemeral port
// that has never been issued in this process AND is still bindable at a
// second moment. Pass it to [appkit.ServiceConfig.Addr] (or
// [NewService], which defaults to its own ephemeral address when cfg.Addr
// is empty) when the test needs to know the address before construction.
func FreeAddr(tb testing.TB) string {
	tb.Helper()

	cfg := &net.ListenConfig{} //nolint:exhaustruct_v5 // zero-value ListenConfig is the documented posture

	for range 25 {
		listener, err := cfg.Listen(context.Background(), "tcp", "localhost:0")
		if err != nil {
			tb.Fatalf("testkit: reserve port: %v", err)
		}

		addr := listener.Addr().String()

		_ = listener.Close()

		freeAddrMu.Lock()

		_, alreadyIssued := freeAddrIssued[addr]
		freeAddrIssued[addr] = struct{}{}

		freeAddrMu.Unlock()

		if alreadyIssued {
			continue
		}

		probe, err := cfg.Listen(context.Background(), "tcp", addr)
		if err != nil {
			continue
		}

		_ = probe.Close()

		return addr
	}

	tb.Fatal("testkit: FreeAddr found no verifiable free port after 25 attempts")

	return ""
}
