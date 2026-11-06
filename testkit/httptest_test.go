package testkit_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/testkit"
)

func TestHTTPGetURLRoundTrip(t *testing.T) {
	t.Parallel()

	svc := testkit.NewService(t, appkit.ServiceConfig{})
	svc.Mux.HandleFunc("GET /kit-get", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("kit says hi"))
	})

	ts := testkit.Serve(t, svc)

	resp := testkit.HTTPGetURL(t, ts.FullChainURL+"/kit-get")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		t.Fatalf("read body: %v", readErr)
	}

	if string(body) != "kit says hi" {
		t.Errorf("body = %q, want %q", string(body), "kit says hi")
	}
}

func TestHTTPGetContextHonorsCancellation(t *testing.T) {
	t.Parallel()

	svc := testkit.NewService(t, appkit.ServiceConfig{})
	svc.Mux.HandleFunc("GET /kit-slow", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
	})

	ts := testkit.Serve(t, svc)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.FullChainURL+"/kit-slow", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, doErr := http.DefaultClient.Do(req)
	if doErr != nil {
		return
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	defer func() { _ = resp.Body.Close() }()

	t.Error("request with canceled context succeeded, want error")
}

func TestWaitForRunningOnServedService(t *testing.T) {
	t.Parallel()

	svc := testkit.NewService(t, appkit.ServiceConfig{})
	testkit.Serve(t, svc)

	testkit.WaitForRunning(t, svc)

	if !svc.Running() {
		t.Error("service not running after WaitForRunning")
	}
}

func TestWaitForRunningWithinExplicitBound(t *testing.T) {
	t.Parallel()

	svc := testkit.NewService(t, appkit.ServiceConfig{})
	testkit.Serve(t, svc)

	testkit.WaitForRunningWithin(t, svc, time.Second)

	if !svc.Running() {
		t.Error("service not running after WaitForRunningWithin")
	}
}

func TestFreeAddrDistinctAndBindable(t *testing.T) {
	t.Parallel()

	first := testkit.FreeAddr(t)
	second := testkit.FreeAddr(t)

	if first == second {
		t.Fatalf("FreeAddr issued the same address twice: %s", first)
	}

	cfg := &net.ListenConfig{}

	probe, err := cfg.Listen(context.Background(), "tcp", first)
	if err != nil {
		t.Fatalf("first FreeAddr address %s not bindable: %v", first, err)
	}

	_ = probe.Close()
}
