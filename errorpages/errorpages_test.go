package errorpages

import (
	"context"
	"encoding/json/v2"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
)

func newMux(t *testing.T) *http.ServeMux {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return mux
}

func TestHandler_FamiliesMapToStatusCodes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		family errorfamily.Family
		want   int
	}{
		{errorfamily.Rejection, http.StatusBadRequest},
		{errorfamily.Conflict, http.StatusConflict},
		{errorfamily.Transient, http.StatusServiceUnavailable},
		{errorfamily.Corruption, http.StatusInternalServerError},
		{errorfamily.Infrastructure, http.StatusServiceUnavailable},
	}

	for _, tc := range cases {
		err := errorfamily.New(tc.family, "test.code", "boom")

		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/boom", nil)

		Handler(err, Config{}).ServeHTTP(rec, req)

		if rec.Code != tc.want {
			t.Errorf("%s: status = %d, want %d", tc.family, rec.Code, tc.want)
		}

		if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
			t.Errorf("%s: content-type = %q, want html", tc.family, ct)
		}
	}
}

func TestHandler_JSONContractShape(t *testing.T) {
	t.Parallel()

	// Transient carries non-empty default Why and Fix guidance.
	err := errorfamily.New(errorfamily.Transient, "store.busy", "database is busy")

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/user", nil)
	req.Header.Set("Accept", "application/json")

	Handler(err, Config{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("content-type = %q, want json", ct)
	}

	var body struct {
		Family  string `json:"family"`
		Code    string `json:"code"`
		Message string `json:"message"`
		Title   string `json:"title"`
		Why     string `json:"why"`
		Fix     string `json:"fix"`
	}

	decodeErr := json.UnmarshalRead(rec.Body, &body)
	if decodeErr != nil {
		t.Fatalf("decode JSON contract: %v", decodeErr)
	}

	if body.Family != "transient" {
		t.Errorf("family = %q, want transient", body.Family)
	}

	if body.Code != "store.busy" {
		t.Errorf("code = %q, want store.busy", body.Code)
	}

	if body.Message == "" {
		t.Error("contract missing message")
	}

	if body.Why == "" || body.Fix == "" {
		t.Errorf("contract missing guidance fields: why=%q fix=%q", body.Why, body.Fix)
	}
}

func TestMount_404HTMLByDefault(t *testing.T) {
	t.Parallel()

	mux := newMux(t)
	Mount(mux, Config{})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/nope", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}

	if !strings.Contains(rec.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("content-type = %q, want html", rec.Header().Get("Content-Type"))
	}

	if !strings.Contains(rec.Body.String(), "404") {
		t.Errorf("body should mention 404, got: %s", rec.Body.String())
	}
}

func TestMount_404JSONWhenAccepted(t *testing.T) {
	t.Parallel()

	mux := newMux(t)
	Mount(mux, Config{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/nope", nil)
	req.Header.Set("Accept", "application/json")

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}

	if !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("content-type = %q, want json", rec.Header().Get("Content-Type"))
	}

	if !strings.Contains(rec.Body.String(), "http.not_found") {
		t.Errorf("body should contain code http.not_found, got: %s", rec.Body.String())
	}
}

func TestMount_RegisteredRoutesUnaffected(t *testing.T) {
	t.Parallel()

	mux := newMux(t)
	Mount(mux, Config{})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestWrap_Pretty404(t *testing.T) {
	t.Parallel()

	handler := Wrap(newMux(t), Config{})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/gone", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}

	if !strings.Contains(rec.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("content-type = %q, want html", rec.Header().Get("Content-Type"))
	}
}

func TestWrap_Pretty405PreservesAllow(t *testing.T) {
	t.Parallel()

	handler := Wrap(newMux(t), Config{})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/health", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}

	if allow := rec.Header().Get("Allow"); !strings.Contains(allow, "GET") {
		t.Errorf("Allow = %q, want GET", allow)
	}

	if !strings.Contains(rec.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("content-type = %q, want html", rec.Header().Get("Content-Type"))
	}
}

func TestWrap_RegisteredRoutesPassThrough(t *testing.T) {
	t.Parallel()

	handler := Wrap(newMux(t), Config{})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestWrite_CustomJSONWhenRule(t *testing.T) {
	t.Parallel()

	err := errorfamily.NewRejection("test.rule", "boom")

	// API path prefix forces JSON even without an Accept header.
	cfg := Config{JSONWhen: func(r *http.Request) bool {
		return strings.HasPrefix(r.URL.Path, "/api/")
	}}

	apiRec := httptest.NewRecorder()
	apiReq := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/thing", nil)
	Write(apiRec, apiReq, err, cfg)

	if !strings.Contains(apiRec.Header().Get("Content-Type"), "application/json") {
		t.Errorf("api path: content-type = %q, want json", apiRec.Header().Get("Content-Type"))
	}

	pageRec := httptest.NewRecorder()
	pageReq := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/page", nil)
	Write(pageRec, pageReq, err, cfg)

	if !strings.Contains(pageRec.Header().Get("Content-Type"), "text/html") {
		t.Errorf("page path: content-type = %q, want html", pageRec.Header().Get("Content-Type"))
	}
}

func TestWrap_FollowsMuxPathCleaningRedirects(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		pattern string // registered pattern; "" registers nothing
		request string
	}{
		{name: "doubled slash on unregistered path", request: "/a//b"},
		{name: "dot segments on unregistered path", request: "/x/../missing"},
		{name: "doubled slash before trailing slash on registered path", pattern: "/health", request: "/health//"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mux := http.NewServeMux()
			if tc.pattern != "" {
				mux.HandleFunc(tc.pattern, func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
			}

			// The bare mux defines expected behavior: a redirect to the
			// cleaned path (307) rather than a 404.
			bareRec := httptest.NewRecorder()
			mux.ServeHTTP(bareRec, httptest.NewRequestWithContext(
				context.Background(), http.MethodGet, "http://example.test"+tc.request, nil))

			if bareRec.Code == http.StatusNotFound {
				t.Fatalf("test setup: bare mux answered 404, want a redirect")
			}

			wrapRec := httptest.NewRecorder()
			Wrap(mux, Config{}).ServeHTTP(wrapRec, httptest.NewRequestWithContext(
				context.Background(), http.MethodGet, "http://example.test"+tc.request, nil))

			if wrapRec.Code != bareRec.Code {
				t.Fatalf("status = %d, want %d (bare mux parity)", wrapRec.Code, bareRec.Code)
			}

			if got, want := wrapRec.Header().Get("Location"), bareRec.Header().Get("Location"); got != want {
				t.Fatalf("Location = %q, want %q (bare mux parity)", got, want)
			}
		})
	}
}

func TestWrap_RendersPretty404OnlyOnCanonicalPath(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()

	// The canonical path a path-cleaning redirect would lead to answers
	// with the pretty 404 page, not another redirect.
	rec := httptest.NewRecorder()
	Wrap(mux, Config{}).ServeHTTP(rec, httptest.NewRequestWithContext(
		context.Background(), http.MethodGet, "http://example.test/missing", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("content-type = %q, want html", ct)
	}
}

// failingWriter simulates a connection that fails on every write (e.g. a
// client disconnecting mid-render). It records the status code the handler
// committed so tests can assert the render-failure fallback path: the
// handler must not panic and must have written the derived status before
// the failing write, so the client still receives the correct code.
type failingWriter struct {
	status  int
	written bool
}

func (f *failingWriter) Header() http.Header { return http.Header{} }

func (f *failingWriter) Write([]byte) (int, error) {
	f.written = true

	return 0, errors.New("connection closed")
}

func (f *failingWriter) WriteHeader(status int) {
	if f.status == 0 {
		f.status = status
	}
}

func TestHandler_RenderFailureFallsBackToPlainStatus(t *testing.T) {
	t.Parallel()

	err := errorfamily.New(errorfamily.Conflict, "cart.version_mismatch", "stale cart")

	fw := &failingWriter{}
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/cart", nil)

	Handler(err, Config{}).ServeHTTP(fw, req)

	if fw.status != http.StatusConflict {
		t.Errorf("committed status = %d, want %d", fw.status, http.StatusConflict)
	}

	if !fw.written {
		t.Error("expected writer to receive at least one write attempt")
	}
}

func TestMount_RenderFailureFallsBackToPlainStatus(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	Mount(mux, Config{})

	fw := &failingWriter{}
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/nowhere", nil)

	mux.ServeHTTP(fw, req)

	if fw.status != http.StatusNotFound {
		t.Errorf("committed status = %d, want %d", fw.status, http.StatusNotFound)
	}

	if !fw.written {
		t.Error("expected writer to receive at least one write attempt")
	}
}
