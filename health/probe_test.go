package health

import (
	"context"
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
