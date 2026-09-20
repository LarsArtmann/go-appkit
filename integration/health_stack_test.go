package integration_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
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
	var trace bytes.Buffer

	recorder, err := fr.New(
		fr.WithWriter(&trace),
		fr.WithMinAge(time.Millisecond),
		fr.WithMaxBytes(1<<20),
	)
	if err != nil {
		t.Fatalf("fr.New: %v", err)
	}
	if err := recorder.Start(); err != nil {
		t.Fatalf("recorder start: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	db := &flakyDependency{}
	injector := do.New()
	do.ProvideNamed(injector, "database", func(do.Injector) (*flakyDependency, error) {
		return db, nil
	})
	if _, err := do.InvokeNamed[*flakyDependency](injector, "database"); err != nil {
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

	startCtx := context.Background()
	if err := mounted.Start(startCtx); err != nil {
		t.Fatalf("mounted start: %v", err)
	}

	shutdown := func(t *testing.T) {
		t.Helper()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := svc.Shutdown(shutdownCtx); err != nil {
			t.Errorf("shutdown: %v", err)
		}
		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("server returned error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("server did not stop after shutdown")
		}
	}
	defer func() {
		if svc.Running() {
			shutdown(t)
		}
	}()

	deadline := time.Now().Add(2 * time.Second)
	for !svc.Running() {
		if time.Now().After(deadline) {
			t.Fatal("service did not start within timeout")
		}
		time.Sleep(time.Millisecond)
	}

	base := "http://" + svc.Addr().String()

	status := func(path string) int {
		t.Helper()

		resp, err := http.Get(base + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		defer resp.Body.Close()
		return resp.StatusCode
	}
	body := func(path string) string {
		t.Helper()

		resp, err := http.Get(base + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		defer resp.Body.Close()
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		return string(raw)
	}

	if code := status("/healthz"); code != http.StatusOK {
		t.Fatalf("/healthz = %d, want 200", code)
	}
	if code := status("/readyz"); code != http.StatusOK {
		t.Fatalf("/readyz = %d, want 200", code)
	}
	if code := status("/health/ready"); code != http.StatusOK {
		t.Fatalf("/health/ready = %d, want 200", code)
	}
	if row := body("/readyz"); !bytes.Contains([]byte(row), []byte("flight-recorder")) {
		t.Fatalf("flight-recorder row missing from /readyz body: %s", row)
	}

	db.failWith(errors.New("connection refused"))

	deadline = time.Now().Add(2 * time.Second)
	for {
		if status("/readyz") == http.StatusServiceUnavailable {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("probe refresh never surfaced the failing database on /readyz")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if code := status("/health/ready"); code != http.StatusServiceUnavailable {
		t.Fatalf("/health/ready = %d, want 503 (cfg.ReadyCheck = mounted.Ready)", code)
	}

	drained := make(chan struct{})
	go func() {
		defer close(drained)
		shutdown(t)
	}()

	deadline = time.Now().Add(2 * time.Second)
	for {
		readyz := status("/readyz")
		appReady := status("/health/ready")
		live := status("/healthz")
		appLive := status("/health/live")
		if readyz == http.StatusServiceUnavailable && appReady == http.StatusServiceUnavailable {
			if live != http.StatusOK || appLive != http.StatusOK {
				t.Fatalf("liveness flipped during drain: /healthz=%d /health/live=%d", live, appLive)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("drain lockstep never observed: /readyz=%d /health/ready=%d", readyz, appReady)
		}
		time.Sleep(5 * time.Millisecond)
	}

	select {
	case <-drained:
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown did not complete")
	}

	recorder.Stop()
	if err := recorder.Close(); err != nil {
		t.Fatalf("recorder close: %v", err)
	}
	if trace.Len() == 0 {
		t.Fatal("trigger captured no trace for the failing batch")
	}
}
