package preset_test

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"

	"github.com/larsartmann/go-appkit/preset"
)

// getWithHeaders fetches url through the full chain with extra headers.
func getWithHeaders(t *testing.T, url string, header map[string]string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	for key, value := range header {
		req.Header.Set(key, value)
	}

	resp, err := (&http.Client{Timeout: httpTimeout}).Do(req)
	if err != nil {
		t.Fatalf("get %s: %v", url, err)
	}

	return resp
}

// postBody posts payload to url with the given Origin header.
func postBody(t *testing.T, url, origin string, payload []byte) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}

	resp, err := (&http.Client{Timeout: httpTimeout}).Do(req)
	if err != nil {
		t.Fatalf("post %s: %v", url, err)
	}

	return resp
}

// registerEcho mounts a body-consuming route on svc.Mux — the body-limit
// middleware only triggers when a handler reads the request body.
func registerEcho(mux interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}) {
	mux.HandleFunc("POST /echo", func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			http.Error(w, "body too large", http.StatusRequestEntityTooLarge)

			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

// TestSecurity_None_DefaultAddsNothing: absent env, nil leg — the standard
// profile's production headers (HSTS) and the body limit never join the
// stack. Core's own basic security headers still apply (nosniff et al).
func TestSecurity_None_DefaultAddsNothing(t *testing.T) {

	svc, err := preset.New(testConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	registerEcho(svc.Mux)

	ts := servePreset(t, svc)

	resp := getWithHeaders(t, ts+"/healthz", nil)
	defer func() { _ = resp.Body.Close() }()

	if resp.Header.Get("Strict-Transport-Security") != "" {
		t.Error("HSTS present with no security leg — the profile silently joined the stack")
	}

	oversize := postBody(t, ts+"/echo", "", bytes.Repeat([]byte("x"), 2<<20))
	defer func() { _ = oversize.Body.Close() }()

	if oversize.StatusCode == http.StatusRequestEntityTooLarge {
		t.Error("413 without the security leg — the body limit silently joined the stack")
	}
}

// TestSecurity_Standard_ProductionHeadersAndBodyLimit: the standard profile
// adds SecurityHeaders (production posture → HSTS) and the 1 MiB body limit.
func TestSecurity_Standard_ProductionHeadersAndBodyLimit(t *testing.T) {

	cfg := testConfig()
	cfg.Security = &preset.SecurityConfig{Enabled: true}

	svc, err := preset.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	registerEcho(svc.Mux)

	ts := servePreset(t, svc)

	page := getWithHeaders(t, ts+"/healthz", nil)
	defer func() { _ = page.Body.Close() }()

	if page.Header.Get("Strict-Transport-Security") == "" {
		t.Error("HSTS missing under the standard profile (production posture)")
	}

	if page.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", page.Header.Get("X-Content-Type-Options"))
	}

	oversize := postBody(t, ts+"/echo", "", bytes.Repeat([]byte("x"), 2<<20))
	defer func() { _ = oversize.Body.Close() }()

	if oversize.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("oversize POST = %d, want 413", oversize.StatusCode)
	}

	// The default limit is 1 MiB: a 64 KiB body passes.
	ok := postBody(t, ts+"/echo", "", bytes.Repeat([]byte("x"), 64<<10))
	defer func() { _ = ok.Body.Close() }()

	if ok.StatusCode != http.StatusOK {
		t.Errorf("64 KiB POST = %d, want 200", ok.StatusCode)
	}
}

// TestSecurity_Strict_OriginCheck: strict adds OriginCheck — cross-origin
// browser requests are rejected with 403, allowlisted origins pass.
func TestSecurity_Strict_OriginCheck(t *testing.T) {

	cfg := testConfig()
	cfg.Security = &preset.SecurityConfig{
		Enabled: true,
		Profile: preset.SecurityStrict,
		Origins: []string{"https://app.example"},
	}

	svc, err := preset.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	registerEcho(svc.Mux)

	ts := servePreset(t, svc)

	evildoer := postBody(t, ts+"/echo", "https://evil.example", []byte("{}"))
	defer func() { _ = evildoer.Body.Close() }()

	if evildoer.StatusCode != http.StatusForbidden {
		t.Errorf("cross-origin POST = %d, want 403", evildoer.StatusCode)
	}

	friend := postBody(t, ts+"/echo", "https://app.example", []byte("{}"))
	defer func() { _ = friend.Body.Close() }()

	if friend.StatusCode != http.StatusOK {
		t.Errorf("allowlisted POST = %d, want 200", friend.StatusCode)
	}
}

// TestSecurity_StrictWithoutOrigins_Rejects: strict with an empty allowlist
// would reject every browser request — New fails fast instead.
func TestSecurity_StrictWithoutOrigins_Rejects(t *testing.T) {

	cfg := testConfig()
	cfg.Security = &preset.SecurityConfig{
		Enabled: true,
		Profile: preset.SecurityStrict,
	}

	_, err := preset.New(cfg)
	if err == nil {
		t.Fatal("New succeeded with strict profile and no origins, want Rejection")
	}

	if !strings.Contains(err.Error(), "Origins") {
		t.Errorf("error %q does not name the missing Origins field", err)
	}

	if errorfamily.Classify(err) != errorfamily.Rejection {
		t.Errorf("error %v classifies as %v, want Rejection", err, errorfamily.Classify(err))
	}
}

// TestSecurity_EnvStandard: APPKIT_SECURITY=standard composes the same
// profile without code.
func TestSecurity_EnvStandard(t *testing.T) {
	t.Setenv("APPKIT_SECURITY", "standard")

	svc, err := preset.New(testConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts := servePreset(t, svc)

	resp := getWithHeaders(t, ts+"/healthz", nil)
	defer func() { _ = resp.Body.Close() }()

	if resp.Header.Get("Strict-Transport-Security") == "" {
		t.Error("HSTS missing with APPKIT_SECURITY=standard")
	}
}
