package integration_test

// httpx × default-stack composition E2E through PUBLISHED tags (httpx +
// core): B4 Conditional behind a live appkit Service — strong-ETag
// revalidation (If-None-Match → 304) through the FULL default middleware
// chain, Server-Timing outermost included. This is httpx's integration
// leg: the module's own suite pins the battery mechanics against
// httptest; here the batteries ride a real service lifecycle.
//
// The ETag header assertion doubles as the default-stack composition pin:
// the 304 must still carry Server-Timing (the outermost middleware stamps
// it at the first write, and a 304 writes headers), proving conditional
// short-circuits compose with the framework stack rather than bypassing it.

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/httpx"
	"github.com/larsartmann/go-appkit/testkit"
)

func TestHTTPxConditionalThroughDefaultStack(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"battery":"b4"}`))
	})

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = testkit.FreeAddr(t)
	cfg.DrainDelay = appkit.NoDrainDelay

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	svc.Mux.Handle("GET /things", httpx.Conditional(handler))

	ts := testkit.Serve(t, svc)
	base := ts.FullChainURL

	waitReady(t, base)

	// First GET: 200 with a strong ETag, through the whole chain.
	first := testkit.HTTPGetURL(t, base+"/things")

	etag := first.Header.Get("ETag")

	if first.StatusCode != http.StatusOK {
		_ = first.Body.Close()
		t.Fatalf("first GET status = %d, want 200", first.StatusCode)
	}

	if etag == "" {
		_ = first.Body.Close()
		t.Fatal("first GET carried no ETag — Conditional is not in the chain")
	}

	_, _ = io.Copy(io.Discard, first.Body)
	_ = first.Body.Close()

	// Revalidation: the same ETag via If-None-Match must short-circuit to
	// 304 without a body, still through the whole chain.
	req, reqErr := http.NewRequestWithContext(
		context.Background(), http.MethodGet, base+"/things", nil)
	if reqErr != nil {
		t.Fatalf("build revalidation request: %v", reqErr)
	}

	req.Header.Set("If-None-Match", etag)

	second, doErr := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if doErr != nil {
		t.Fatalf("revalidation GET: %v", doErr)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, second.Body)
		_ = second.Body.Close()
	}()

	if second.StatusCode != http.StatusNotModified {
		t.Fatalf("revalidation status = %d, want 304 (If-None-Match: %s)", second.StatusCode, etag)
	}

	if st := second.Header.Get("Server-Timing"); st == "" {
		t.Error("304 carried no Server-Timing header — the conditional short-circuit bypassed the default stack")
	}
}
