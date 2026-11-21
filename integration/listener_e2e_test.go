package integration_test

// listener-seam × published-tag E2E (SUPERB v17 T6): a service on a
// PRE-BOUND unix socket through the FULL default middleware chain — the
// consumer shape both waiting reviews named (go-compliance: "appkit core
// binds TCP only"; nsfw-classifier: "grep WithListener: zero hits"). The
// module's own suite pins the seam mechanics (ownership transfer, failure
// paths); here the seam rides a real Service lifecycle on a published tag,
// including the ownership pin: graceful shutdown closes the SUPPLIED
// listener and Addr() reports the socket path.

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/testkit"
)

func TestListenerUnixSocketThroughPublishedTags(t *testing.T) {
	t.Parallel()

	socketPath := filepath.Join(t.TempDir(), "appkit-integration.sock")

	var lc net.ListenConfig

	listener, err := lc.Listen(t.Context(), "unix", socketPath)
	if err != nil {
		t.Fatalf("unix bind: %v", err)
	}

	cfg := appkit.DefaultServiceConfig()
	cfg.Listener = listener // Addr is ignored; the service serves the socket
	cfg.DrainDelay = appkit.NoDrainDelay

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	testkit.WaitForRunning(t, svc)

	if got := svc.Addr().String(); got != socketPath {
		t.Errorf("Addr() = %q, want socket path %q", got, socketPath)
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var d net.Dialer

				return d.DialContext(ctx, "unix", socketPath)
			},
		},
	}

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://unix/health", nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unix-dialed health check: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health over unix socket = %d, want 200 (body: %s)", resp.StatusCode, body)
	}

	var health struct {
		Status string `json:"status"`
	}

	if err := json.Unmarshal(body, &health); err != nil {
		t.Fatalf("decode health: %v", err)
	}

	if health.Status != "up" {
		t.Errorf("health status = %q, want %q", health.Status, "up")
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancelShutdown()

	if err := svc.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("serve error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not stop")
	}

	// Ownership pin through the published tag: the SUPPLIED listener closed.
	dialCtx, cancelDial := context.WithTimeout(t.Context(), time.Second)
	defer cancelDial()

	var d net.Dialer

	if _, err := d.DialContext(dialCtx, "unix", socketPath); err == nil {
		t.Error("supplied unix listener still accepting after shutdown; want closed")
	}

	if _, statErr := os.Stat(socketPath); !os.IsNotExist(statErr) {
		t.Errorf("socket file still present after shutdown (stat err = %v)", statErr)
	}
}
