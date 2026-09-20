package integration_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit"
	frhealth "github.com/larsartmann/go-appkit/flightrecorderhealth"
	appkithealth "github.com/larsartmann/go-appkit/health"
	fr "github.com/larsartmann/go-flightrecorder"
	health "github.com/larsartmann/go-health"
	"github.com/samber/do/v2"
)

// flakyDependency is a named injector service whose failure can be flipped
// at test runtime.
type flakyDependency struct {
	mu  sync.Mutex
	err error
}

func (f *flakyDependency) HealthCheck(_ context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.err
}

func (f *flakyDependency) failWith(err error) {
	f.mu.Lock()
	f.err = err
	f.mu.Unlock()
}

// stackProbe is the HTTP voice of the running service under test: status
// codes, response bodies, and deadline-bounded condition waits, all
// fatal-on-error into the test.
type stackProbe struct {
	t     *testing.T
	fetch func(path string) (int, string)
}

func newStackProbe(t *testing.T, base string) *stackProbe {
	t.Helper()

	if code := healthProbe.status("/healthz"); code != http.StatusOK {
		t.Fatalf("/healthz = %d, want 200", code)
	}
	if code := healthProbe.status("/readyz"); code != http.StatusOK {
		t.Fatalf("/readyz = %d, want 200", code)
	}
	if code := healthProbe.status("/health/ready"); code != http.StatusOK {
		t.Fatalf("/health/ready = %d, want 200", code)
	}
	if _, body := healthProbe.fetch("/readyz"); !strings.Contains(body, "flight-recorder") {
		t.Fatalf("flight-recorder row missing from /readyz body: %s", body)
	}

	db.failWith(errors.New("connection refused"))

	// Both readiness surfaces must report the failing critical dependency:
	// the module's via the probe refresh loop, appkit's via ReadyCheck.
	healthProbe.waitFor("failing database on both readiness surfaces", func() bool {
		return healthProbe.status("/readyz") == http.StatusServiceUnavailable &&
			healthProbe.status("/health/ready") == http.StatusServiceUnavailable
	})

	drained := make(chan struct{})
	go func() {
		defer close(drained)

		shutdown()
	}()

	// During the drain window both readiness surfaces must read 503 in
	// lockstep while both liveness surfaces stay 200. All four checks run
	// inside the observer so no assertion can race past server stop.
	drainLockstep := func() bool {
		readyz := healthProbe.status("/readyz")
		appReady := healthProbe.status("/health/ready")
		if readyz != http.StatusServiceUnavailable || appReady != http.StatusServiceUnavailable {
			return false
		}

		live := healthProbe.status("/healthz")
		appLive := healthProbe.status("/health/live")
		if live != http.StatusOK || appLive != http.StatusOK {
			t.Fatalf("liveness flipped during drain: /healthz=%d /health/live=%d", live, appLive)
		}

		return true
	}
	healthProbe.waitFor("drain lockstep on both readiness surfaces", drainLockstep)

	select {
	case <-drained:
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown did not complete")
	}

	recorder.Stop()
	if trace.Len() == 0 {
		t.Fatal("trigger captured no trace for the failing batch")
	}
}
