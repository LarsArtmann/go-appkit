package security_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/larsartmann/go-appkit/security"
)

func originResponse(middleware func(http.Handler) http.Handler, headers map[string]string) *httptest.ResponseRecorder {
	var reached bool
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		w.WriteHeader(http.StatusTeapot) // 418: unmistakably "handler ran"
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/form", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !reached && rec.Code == http.StatusTeapot {
		t.Error("handler reached flag lost")
	}

	return rec
}

func TestOriginCheck_SameOriginAllowed(t *testing.T) {
	t.Parallel()

	mw := security.OriginCheck([]string{"https://app.example.com"}, nil)

	req1 := originResponse(mw, map[string]string{"Origin": "http://localhost:8080"})
	if req1.Code != http.StatusTeapot {
		t.Errorf("same-origin POST (no allowlist entry): status = %d, want 418 — browsers attach Origin to same-origin fetches too", req1.Code)
	}

	req2 := originResponse(mw, map[string]string{"Origin": "https://app.example.com"})
	if req2.Code != http.StatusTeapot {
		t.Errorf("allowlisted origin: status = %d, want 418", req2.Code)
	}
}

func TestOriginCheck_CrossOriginRejected(t *testing.T) {
	t.Parallel()

	mw := security.OriginCheck([]string{"https://app.example.com"}, nil)

	rec := originResponse(mw, map[string]string{"Origin": "https://evil.example.net"})
	if rec.Code != http.StatusForbidden {
		t.Errorf("cross-origin: status = %d, want 403", rec.Code)
	}
}

func TestOriginCheck_NoOriginOrRefererPasses(t *testing.T) {
	t.Parallel()

	mw := security.OriginCheck([]string{"https://app.example.com"}, nil)

	rec := originResponse(mw, nil)
	if rec.Code != http.StatusTeapot {
		t.Errorf("no Origin/Referer (curl, machine client): status = %d, want 418", rec.Code)
	}
}

func TestOriginCheck_RefererFallback(t *testing.T) {
	t.Parallel()

	mw := security.OriginCheck([]string{"https://app.example.com"}, nil)

	rec := originResponse(mw, map[string]string{"Referer": "https://app.example.com/form"})
	if rec.Code != http.StatusTeapot {
		t.Errorf("Referer fallback to allowlisted origin: status = %d, want 418", rec.Code)
	}

	rec2 := originResponse(mw, map[string]string{"Referer": "https://evil.example.net/form"})
	if rec2.Code != http.StatusForbidden {
		t.Errorf("Referer fallback to foreign origin: status = %d, want 403", rec2.Code)
	}
}

func TestOriginCheck_CaseInsensitiveMatch(t *testing.T) {
	t.Parallel()

	mw := security.OriginCheck([]string{"https://APP.example.com"}, nil)

	rec := originResponse(mw, map[string]string{"Origin": "https://app.example.com"})
	if rec.Code != http.StatusTeapot {
		t.Errorf("case-insensitive origin: status = %d, want 418", rec.Code)
	}
}

func TestOriginCheck_AllowAllLogs(t *testing.T) {
	t.Parallel()

	capture := &recordingHandler{}
	mw := security.OriginCheck([]string{"*"}, slog.New(capture))

	rec := originResponse(mw, map[string]string{"Origin": "https://anything.example"})
	if rec.Code != http.StatusTeapot {
		t.Errorf("allow-all: status = %d, want 418", rec.Code)
	}

	if !capture.contains("allow-all") {
		t.Error("allow-all configuration must log a warning at construction")
	}
}

// recordingHandler is a minimal slog.Handler capturing message strings.
type recordingHandler struct {
	mu  sync.Mutex
	msg []string
}

func (h *recordingHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *recordingHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.msg = append(h.msg, r.Message)

	return nil
}

func (h *recordingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h *recordingHandler) WithGroup(string) slog.Handler { return h }

func (h *recordingHandler) contains(substr string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, m := range h.msg {
		if strings.Contains(m, substr) {
			return true
		}
	}

	return false
}
