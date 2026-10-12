package health

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/go-health"
)

func TestNewProbe_AllHealthyReportsPass(t *testing.T) {
	t.Parallel()

	probe := NewProbe(map[string]CheckFunc{
		"database": func(context.Context) error { return nil },
		"cache":    func(context.Context) error { return nil },
	})

	err := probe.Start(t.Context())
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	if got := probe.Status(); got != health.StatusPass {
		t.Errorf("status = %v, want pass", got)
	}

	if !probe.Ready() {
		t.Error("probe with all checks passing must report ready")
	}
}

func TestNewProbe_ClassificationFollowsCriticality(t *testing.T) {
	t.Parallel()

	failing := map[string]CheckFunc{
		"database": func(context.Context) error { return errors.New("connection refused") },
		"cache":    func(context.Context) error { return errors.New("cache down") },
	}

	t.Run("critical failure fails readiness", func(t *testing.T) {
		t.Parallel()

		probe := NewProbe(failing, health.WithCriticalServices("database"))

		err := probe.Start(t.Context())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		if got := probe.Status(); got != health.StatusFail {
			t.Errorf("status = %v, want fail", got)
		}

		if probe.Ready() {
			t.Error("critical failure must clear readiness")
		}
	})

	t.Run("non-critical failure degrades to warn", func(t *testing.T) {
		t.Parallel()

		probe := NewProbe(map[string]CheckFunc{
			"database": func(context.Context) error { return nil },
			"cache":    failing["cache"],
		})

		err := probe.Start(t.Context())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		if got := probe.Status(); got != health.StatusWarn {
			t.Errorf("status = %v, want warn", got)
		}

		if !probe.Ready() {
			t.Error("non-critical failure must keep readiness")
		}
	})
}

func TestNewProbe_PanicIsolatedPerCheck(t *testing.T) {
	t.Parallel()

	probe := NewProbe(map[string]CheckFunc{
		"database": func(context.Context) error { return nil },
		"cache":    func(context.Context) error { panic("cache exploded") },
	})

	err := probe.Start(t.Context())
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	resp := probe.CachedResponse()

	cache, ok := resp.Checks["cache"]
	if !ok {
		t.Fatal("panicking check missing from response")
	}

	if !strings.Contains(cache.Error, "panicked") {
		t.Errorf("check error = %q, want panic report", cache.Error)
	}

	if got := resp.Status; got != health.StatusWarn {
		t.Errorf("status = %v, want warn (non-critical panic degrades, not fails)", got)
	}
}

func TestNewProbe_CriticalPanicFailsReadiness(t *testing.T) {
	t.Parallel()

	probe := NewProbe(map[string]CheckFunc{
		"database": func(context.Context) error { panic("db exploded") },
	}, health.WithCriticalServices("database"))

	err := probe.Start(t.Context())
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	if got := probe.Status(); got != health.StatusFail {
		t.Errorf("status = %v, want fail", got)
	}
}

func TestNewProbe_ChecksRunConcurrently(t *testing.T) {
	t.Parallel()

	firstStarted := make(chan struct{})
	secondStarted := make(chan struct{})

	probe := NewProbe(map[string]CheckFunc{
		"first": func(context.Context) error {
			close(firstStarted)

			select {
			case <-secondStarted:
			case <-time.After(testTimeout):
				t.Error("second check never started; checks are not concurrent")
			}

			return nil
		},
		"second": func(context.Context) error {
			close(secondStarted)

			select {
			case <-firstStarted:
			case <-time.After(testTimeout):
				t.Error("first check never started; checks are not concurrent")
			}

			return nil
		},
	})

	err := probe.Start(t.Context())
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	if !probe.Ready() {
		t.Error("concurrent handshake checks must both pass")
	}
}

func TestNewProbe_ChecksReceiveBoundedContext(t *testing.T) {
	t.Parallel()

	sawDeadline := false

	probe := NewProbe(map[string]CheckFunc{
		"database": func(ctx context.Context) error {
			_, ok := ctx.Deadline()
			sawDeadline = ok

			return nil
		},
	})

	err := probe.Start(t.Context())
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	if !sawDeadline {
		t.Error("checks must receive a timeout-bounded context")
	}
}

// TestNewProbe_ReportsPerCheckDurations pins the metadata fix (#25): every
// check is self-timed, so the served response carries per-check
// duration_ns (previously always omitted on the NewProbe path). The slow
// check's duration must reflect its runtime, and the field name must
// survive JSON serialization — that is what duration-rendering dashboards
// consume.
func TestNewProbe_ReportsPerCheckDurations(t *testing.T) {
	t.Parallel()

	probe := NewProbe(map[string]CheckFunc{
		"fast": func(context.Context) error { return nil },
		"slow": func(ctx context.Context) error {
			select {
			case <-time.After(30 * time.Millisecond):
			case <-ctx.Done():
				t.Error("batch deadline cut the slow check short")
			}

			return nil
		},
	})

	resp := probe.Evaluate(t.Context())

	slow, ok := resp.Checks["slow"]
	if !ok {
		t.Fatal("slow check missing from response")
	}

	if slow.DurationNanos < 25*time.Millisecond.Nanoseconds() {
		t.Errorf("slow check duration_ns = %d, want >=25ms (check must be self-timed)", slow.DurationNanos)
	}

	fast, ok := resp.Checks["fast"]
	if !ok {
		t.Fatal("fast check missing from response")
	}

	if fast.DurationNanos <= 0 {
		t.Errorf("fast check duration_ns = %d, want >0 (every check reports timing)", fast.DurationNanos)
	}

	wire, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	if !bytes.Contains(wire, []byte(`"duration_ns"`)) {
		t.Errorf("wire response omits duration_ns: %s", wire)
	}
}

// TestNewProbe_WedgedCheckFailsClosedAtDeadline pins the fail-closed
// guarantee the NewChecks switch adds: a check that ignores its context is
// abandoned at the batch deadline with an error naming it — the batch
// completes, other checks report, and one wedged dependency cannot hang
// the probe. The abandoned check goroutine eventually exits on its own
// (time.After below); nothing blocks on it.
func TestNewProbe_WedgedCheckFailsClosedAtDeadline(t *testing.T) {
	t.Parallel()

	probe := NewProbe(map[string]CheckFunc{
		"healthy": func(context.Context) error { return nil },
		"wedged": func(context.Context) error {
			time.Sleep(2 * time.Second) // deliberate ctx-ignoring wedged check

			return nil
		},
	}, health.WithTimeout(250*time.Millisecond))

	err := probe.Start(t.Context())
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	resp := probe.CachedResponse()

	wedged, ok := resp.Checks["wedged"]
	if !ok {
		t.Fatal("wedged check missing from response")
	}

	if !strings.Contains(wedged.Error, "did not finish before the batch deadline") {
		t.Errorf("wedged check error = %q, want fail-closed deadline report", wedged.Error)
	}

	if got := resp.Status; got != health.StatusWarn {
		t.Errorf("status = %v, want warn (non-critical wedged check degrades, batch completes)", got)
	}

	healthy, ok := resp.Checks["healthy"]
	if !ok || healthy.Error != "" {
		t.Errorf("healthy check must still pass: ok=%v check=%+v", ok, healthy)
	}
}
