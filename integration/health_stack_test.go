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

// newHealthStackService wires the full composed stack against published
// tags: fr.Recorder (trace buffer), injector with a flaky "database" and the
// flight-recorder Checkable, the injector-path probe with the Trigger
// recorder, and a Mounted surface on a fresh appkit Service with drain
// lockstep wired (DrainHooks/ShutdownHooks/ReadyCheck). The recorder's
// async captures are drained by recorder.Stop; Close is cleanup-managed.
func newHealthStackService(
	t *testing.T,
) (*appkit.Service, <-chan error, *flakyDependency, *appkithealth.Mounted, *fr.Recorder, *bytes.Buffer) {
	t.Helper()

	var trace bytes.Buffer

	recorder, err := fr.New(
		fr.WithWriter(&trace),
		fr.WithMinAge(time.Millisecond),
		fr.WithMaxBytes(1<<20),
	)
	if err != nil {
		t.Fatalf("fr.New: %v", err)
	}
	err = recorder.Start()
	if err != nil {
		t.Fatalf("recorder start: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	db := &flakyDependency{}
	injector := do.New()
	do.ProvideNamed(injector, "database", func(do.Injector) (*flakyDependency, error) {
		return db, nil
	})
	_, err = do.InvokeNamed[*flakyDependency](injector, "database")
	if err != nil {
		t.Fatalf("invoke database: %v", err)
	}
	checkable := frhealth.Register(injector, recorder, "flight-recorder")
	_ = checkable

	probe := health.New(injector,
		health.WithHealthRecorder(frhealth.NewTrigger(recorder)),
		health.WithCriticalServices("database"),
		health.WithRefreshInterval(50*time.Millisecond))

	mounted, err := appkithealth.New(probe)
	if err != nil {
		t.Fatalf("appkithealth.New: %v", err)
	}

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = freeAddr(t)
	cfg.DrainDelay = 250 * time.Millisecond
	cfg.DrainHooks = append(cfg.DrainHooks, func(context.Context) error {
		mounted.Drain()

		return nil
	})
	cfg.ShutdownHooks = append(cfg.ShutdownHooks, mounted.Shutdown)
	cfg.ReadyCheck = mounted.Ready

	svc, err := appkit.NewService(cfg)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	mounted.RegisterRoutes(svc.Mux)

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start service: %v", err)
	}

	started := time.Now().Add(2 * time.Second)
	for !svc.Running() {
		if time.Now().After(started) {
			t.Fatal("service did not start within timeout")
		}

		time.Sleep(time.Millisecond)
	}

	t.Cleanup(func() {
		err := recorder.Close()
		if err != nil {
			t.Errorf("recorder close: %v", err)
		}
	})

	return svc, errCh, db, mounted, recorder, &trace
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

	fetch := func(path string) (int, string) {
		t.Helper()

		reqCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, base+path, nil)
		if err != nil {
			t.Fatalf("request %s: %v", path, err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		defer func() { _ = resp.Body.Close() }()
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}

		return resp.StatusCode, string(raw)
	}

	return &stackProbe{t: t, fetch: fetch}
}

func (s *stackProbe) fetchBody(path string) string {
	s.t.Helper()

	_, body := s.fetch(path)

	return body
}

func (s *stackProbe) status(path string) int {
	s.t.Helper()

	code, _ := s.fetch(path)

	return code
}

func (s *stackProbe) waitFor(desc string, observed func() bool) {
	s.t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for !observed() {
		if time.Now().After(deadline) {
			s.t.Fatalf("%s never observed", desc)
		}

		time.Sleep(5 * time.Millisecond)
	}
}

// TestHealthStackThroughAppkitService pins the full self-health composition
// end to end against PUBLISHED tags: a samber/do injector registers a
// dependency plus the flightrecorder Checkable, go-health's injector-path
// probe honors WithHealthRecorder (frhealth.Trigger), and appkithealth's
// Mounted surface rides the Service's drain window in lockstep with the
// framework's own ready probe (cfg.ReadyCheck = mounted.Ready composes the
// dependency failure into /health/ready before the drain).
//
// Asserted, in order: all readiness surfaces 200 with the flight-recorder
// row visible; a failing critical dependency surfaces as 503 on BOTH the
// module's /readyz and appkit's /health/ready; during the svc.Shutdown
// drain window both surfaces stay 503 while both liveness surfaces stay
// 200; the Trigger captured a real runtime trace for the failing batch.
func TestHealthStackThroughAppkitService(t *testing.T) {
	t.Parallel()

	svc, errCh, db, mounted, recorder, trace := newHealthStackService(t)

	shutdown := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownErr := svc.Shutdown(shutdownCtx)
		if shutdownErr != nil {
			t.Errorf("shutdown: %v", shutdownErr)
		}
		select {
		case serveErr := <-errCh:
			if serveErr != nil {
				t.Errorf("server returned error: %v", serveErr)
			}
		case <-time.After(2 * time.Second):
			t.Error("server did not stop after shutdown")
		}
	}
	defer func() {
		if svc.Running() {
			shutdown()
		}
	}()

	startCtx := context.Background()
	err := mounted.Start(startCtx)
	if err != nil {
		t.Fatalf("mounted start: %v", err)
	}

	healthProbe := newStackProbe(t, "http://"+svc.Addr().String())

	if code := healthProbe.status("/healthz"); code != http.StatusOK {
		t.Fatalf("/healthz = %d, want 200", code)
	}
	if code := healthProbe.status("/readyz"); code != http.StatusOK {
		t.Fatalf("/readyz = %d, want 200", code)
	}
	if code := healthProbe.status("/health/ready"); code != http.StatusOK {
		t.Fatalf("/health/ready = %d, want 200", code)
	}
	if body := healthProbe.fetchBody("/readyz"); !strings.Contains(body, "flight-recorder") {
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
