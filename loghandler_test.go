package appkit

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"strings"
	"testing"
)

func TestLogHandler_ReceivesLifecycleLines(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	handler := slog.NewTextHandler(
		&buf,
		&slog.HandlerOptions{Level: slog.LevelInfo},
	)

	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		LogHandler: handler,
		DrainDelay: NoDrainDelay,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if svc.Logger.Handler() != handler {
		t.Fatal("service logger must emit through the supplied slog.Handler")
	}

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForRunning(t, svc)

	ctx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	resp := httpGet(t, ctx, "http://"+svc.Addr().String()+"/health")
	drainAndClose(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("health through handler-swapped service = %d, want 200", resp.StatusCode)
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(t.Context(), testTimeout)
	defer cancelShutdown()

	if err := svc.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	assertServerStopped(t, errCh)

	logs := buf.String()

	for _, want := range []string{
		"shutdown phase complete",
		"phase=ready_flip",
		"phase=shutdown_hooks",
		"graceful shutdown complete",
	} {
		if !strings.Contains(logs, want) {
			t.Errorf("supplied-handler logs missing %q; logs:\n%s", want, logs)
		}
	}
}

func TestLogHandler_LogLevelAndFormatIgnored(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	handler := slog.NewTextHandler(
		&buf,
		&slog.HandlerOptions{Level: slog.LevelInfo},
	)

	// Values that would FAIL InitLogger (unsupported level/format) must not
	// matter when a handler is supplied: the handler owns level+format.
	svc, err := NewService(ServiceConfig{
		Addr:       "localhost:0",
		LogLevel:   LogLevel("bogus"),
		LogFormat:  LogFormat("bogus"),
		LogHandler: handler,
	})
	if err != nil {
		t.Fatalf("new service with bogus level/format + handler: %v", err)
	}

	if svc.Logger == nil {
		t.Fatal("service logger must exist")
	}

	shutdownCtx, cancel := context.WithTimeout(t.Context(), testTimeout)
	defer cancel()

	// Never started: Shutdown is a no-op and logs nothing.
	if err := svc.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown of never-started service: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("never-started service must log nothing, got:\n%s", buf.String())
	}
}
