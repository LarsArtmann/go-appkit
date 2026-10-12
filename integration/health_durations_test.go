package integration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	appkit "github.com/larsartmann/go-appkit"
	appkithealth "github.com/larsartmann/go-appkit/health"
	"github.com/larsartmann/go-appkit/testkit"
	healthsdk "github.com/larsartmann/go-health"
)

// TestHealthDurationsThroughAppkitService pins health v0.2.0's
// NewProbe->NewChecks delta through the FULL appkit stack on PUBLISHED tags
// (core v0.9.0 + health v0.2.0): per-check duration_ns must survive the
// default middleware chain to the /readyz wire, and a check that ignores its
// context must fail closed at the batch deadline instead of hanging the
// readiness answer — issue #25's two halves, composed end-to-end. The
// module-level unit tests (health/probe_test.go) pin the probe in isolation;
// this leg proves the composition a consumer actually runs.
func TestHealthDurationsThroughAppkitService(t *testing.T) {
	t.Parallel()

	probe := appkithealth.NewProbe(map[string]appkithealth.CheckFunc{
		"fast": func(context.Context) error { return nil },
		"slow": func(ctx context.Context) error {
			select {
			case <-time.After(30 * time.Millisecond):
			case <-ctx.Done():
				t.Error("batch deadline cut the slow check short")
			}

			return nil
		},
		"wedged": func(context.Context) error {
			time.Sleep(2 * time.Second) // deliberately ignores its context

			return nil
		},
	}, healthsdk.WithTimeout(250*time.Millisecond))

	mounted, err := appkithealth.New(probe)
	if err != nil {
		t.Fatalf("appkithealth.New: %v", err)
	}

	cfg := appkit.DefaultServiceConfig()
	cfg.DrainHooks = append(cfg.DrainHooks, func(context.Context) error {
		mounted.Drain()

		return nil
	})
	cfg.ShutdownHooks = append(cfg.ShutdownHooks, mounted.Shutdown)
	cfg.ReadyCheck = mounted.Ready

	svc := testkit.NewService(t, cfg)
	mounted.RegisterRoutes(svc.Mux)

	ctx := context.Background()
	err = mounted.Start(ctx)
	if err != nil {
		t.Fatalf("mounted.Start: %v", err)
	}

	ts := testkit.Serve(t, svc)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.FullChainURL+"/readyz", nil)
	if err != nil {
		t.Fatalf("build /readyz request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /readyz: %v", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read /readyz body: %v", err)
	}

	// Non-critical wedged check degrades to warn: readiness stays 200.
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/readyz = %d, want 200 (wedged check is non-critical): %s", resp.StatusCode, body)
	}

	if !strings.Contains(string(body), `"duration_ns"`) {
		t.Fatalf("wire response omits duration_ns through the appkit stack: %s", body)
	}

	var payload struct {
		Status string `json:"status"`
		Checks map[string]struct {
			Status        string `json:"status"`
			Error         string `json:"error"`
			DurationNanos int64  `json:"duration_ns"` //nolint:tagliatelle // the wire contract is go-health's snake_case field, not this repo's camelCase
		} `json:"checks"`
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		t.Fatalf("unmarshal /readyz body: %v: %s", err, body)
	}

	slow, ok := payload.Checks["slow"]
	if !ok {
		t.Fatalf("slow check missing from wire response: %s", body)
	}
	if slow.DurationNanos < 25*time.Millisecond.Nanoseconds() {
		t.Errorf("slow check duration_ns = %d, want >=25ms (self-timed through the full stack)", slow.DurationNanos)
	}

	fast, ok := payload.Checks["fast"]
	if !ok {
		t.Fatalf("fast check missing from wire response: %s", body)
	}
	if fast.DurationNanos <= 0 {
		t.Errorf("fast check duration_ns = %d, want >0", fast.DurationNanos)
	}

	wedged, ok := payload.Checks["wedged"]
	if !ok {
		t.Fatalf("wedged check missing from wire response: %s", body)
	}
	if !strings.Contains(wedged.Error, "did not finish before the batch deadline") {
		t.Errorf("wedged check error = %q, want fail-closed deadline report through the full stack", wedged.Error)
	}
	if want := string(healthsdk.StatusWarn); payload.Status != want {
		t.Errorf("roll-up status = %q, want %q (non-critical wedged check degrades, batch completes)", payload.Status, want)
	}
}
