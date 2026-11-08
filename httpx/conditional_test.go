package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// feedHandler is the canonical conditional consumer: a deterministic body,
// a known modification time, stamped through SetLastModified.
func feedHandler(modifiedAt time.Time, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		SetLastModified(w, modifiedAt)
		w.Header().Set("Content-Type", "text/calendar")
		_, _ = w.Write([]byte(body))
	})
}

func TestConditional_StrongETagAndRevalidation(t *testing.T) {
	t.Parallel()

	modified := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(Conditional(feedHandler(modified, "BEGIN:VCALENDAR\nEND:VCALENDAR\n")))
	defer srv.Close()

	first, err := http.Get(srv.URL) //nolint:noctx // test server probe with implicit context
	if err != nil {
		t.Fatalf("first GET: %v", err)
	}
	defer first.Body.Close()

	if first.StatusCode != http.StatusOK {
		t.Fatalf("first GET status = %d, want 200", first.StatusCode)
	}

	etag := first.Header.Get("ETag")
	if etag == "" || strings.HasPrefix(etag, "W/") {
		t.Fatalf("ETag = %q, want a non-empty strong tag", etag)
	}

	if first.Header.Get("Last-Modified") == "" {
		t.Fatal("Last-Modified missing — the handler's stamp did not ride through")
	}

	// The revalidation: same tag → 304 with no body.
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	req.Header.Set("If-None-Match", etag)

	second, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("revalidation GET: %v", err)
	}
	defer second.Body.Close()

	if second.StatusCode != http.StatusNotModified {
		t.Fatalf("revalidation status = %d, want 304", second.StatusCode)
	}

	if second.Header.Get("ETag") != etag {
		t.Fatalf("304 ETag = %q, want the matching validator kept", second.Header.Get("ETag"))
	}
}

// The CV conditional matrix (TestInterviewsICSHandler_ConditionalGET):
// weak forms, comma lists, and * must all answer 304.
func TestConditional_IfNoneMatchMatrix(t *testing.T) {
	t.Parallel()

	modified := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(Conditional(feedHandler(modified, "stable-body")))
	t.Cleanup(srv.Close)

	first, err := http.Get(srv.URL) //nolint:noctx // test server probe
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer first.Body.Close()

	etag := first.Header.Get("ETag")

	cases := []struct {
		name  string
		value string
	}{
		{"exact strong", etag},
		{"weak form", "W/" + etag},
		{"comma list", `"deadbeef", ` + etag},
		{"star", "*"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
			req.Header.Set("If-None-Match", tc.value)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("GET: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusNotModified {
				t.Fatalf("If-None-Match %q status = %d, want 304", tc.value, resp.StatusCode)
			}
		})
	}
}

func TestConditional_IfModifiedSince(t *testing.T) {
	t.Parallel()

	modified := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(Conditional(feedHandler(modified, "stable-body")))
	t.Cleanup(srv.Close)

	cases := []struct {
		name string
		since time.Time
		want  int
	}{
		{"client copy current", modified, http.StatusNotModified},
		{"client copy newer", modified.Add(time.Hour), http.StatusNotModified},
		{"client copy stale", modified.Add(-time.Hour), http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
			req.Header.Set("If-Modified-Since", tc.since.UTC().Format(http.TimeFormat))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("GET: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.want)
			}
		})
	}
}

// RFC precedence: when If-None-Match is present, If-Modified-Since is
// ignored — a stale time header must not resurrect a response the (fresh)
// ETag says is modified, and a fresh ETag must win even when the time
// validator would say 200.
func TestConditional_IfNoneMatchTakesPrecedenceOverIMS(t *testing.T) {
	t.Parallel()

	modified := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(Conditional(feedHandler(modified, "stable-body")))
	t.Cleanup(srv.Close)

	first, err := http.Get(srv.URL) //nolint:noctx // test server probe
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer first.Body.Close()

	etag := first.Header.Get("ETag")

	// Stale copy by TIME (would be 200), but the ETag matches → 304 by
	// If-None-Match precedence.
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	req.Header.Set("If-None-Match", etag)
	req.Header.Set("If-Modified-Since", modified.Add(-time.Hour).UTC().Format(http.TimeFormat))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotModified {
		t.Fatalf("status = %d, want 304 (If-None-Match decides, IMS ignored)", resp.StatusCode)
	}
}

// A mismatching If-None-Match with a current If-Modified-Since: the ETag
// says modified, so the response is 200 even though the time validator
// alone would have said 304.
func TestConditional_MismatchedETagBeatsCurrentIMS(t *testing.T) {
	t.Parallel()

	modified := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(Conditional(feedHandler(modified, "stable-body")))
	t.Cleanup(srv.Close)

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	req.Header.Set("If-None-Match", `"totally-different"`)
	req.Header.Set("If-Modified-Since", modified.UTC().Format(http.TimeFormat))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (mismatched If-None-Match decides: modified)", resp.StatusCode)
	}
}

func TestConditional_HeadKeepsValidatorsNoBody(t *testing.T) {
	t.Parallel()

	modified := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(Conditional(feedHandler(modified, "stable-body")))
	t.Cleanup(srv.Close)

	get, err := http.Get(srv.URL) //nolint:noctx // test server probe
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer get.Body.Close()

	etag := get.Header.Get("ETag")

	head, err := http.NewRequestWithContext(t.Context(), http.MethodHead, srv.URL, nil)
	if err != nil {
		t.Fatalf("build HEAD: %v", err)
	}

	head.Header.Set("If-None-Match", etag)

	headResp, err := http.DefaultClient.Do(head)
	if err != nil {
		t.Fatalf("HEAD: %v", err)
	}
	defer headResp.Body.Close()

	if headResp.StatusCode != http.StatusNotModified {
		t.Fatalf("HEAD status = %d, want 304", headResp.StatusCode)
	}
}

func TestConditional_SkipBypassesValidators(t *testing.T) {
	t.Parallel()

	handler := Conditional(
		feedHandler(time.Now(), "stable-body"),
		SkipConditional(func(r *http.Request) bool { return strings.HasSuffix(r.URL.Path, "/stream") }),
	)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+"/stream", nil)
	req.Header.Set("If-None-Match", "*")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /stream: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("skipped request status = %d, want an untouched 200", resp.StatusCode)
	}

	if resp.Header.Get("ETag") != "" {
		t.Fatal("skipped request carries an ETag — SkipConditional did not bypass the battery")
	}
}

func TestConditional_ETagFuncOverride(t *testing.T) {
	t.Parallel()

	handler := Conditional(
		feedHandler(time.Now(), "stable-body"),
		ETagFunc(func(body []byte) string { return "custom-opaque" }),
	)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL) //nolint:noctx // test server probe
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("ETag"); got != `"custom-opaque"` {
		t.Fatalf("ETag = %q, want the custom opaque tag quoted", got)
	}
}

func TestConditional_NoValidatorsSentPlain200(t *testing.T) {
	t.Parallel()

	handler := Conditional(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("body"))
	}))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	// No ETag match, no IMS → plain 200 even with a garbage IMS header.
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	req.Header.Set("If-Modified-Since", "not-a-date")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (unparseable IMS is ignored)", resp.StatusCode)
	}
}

func TestSetLastModified_Format(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	SetLastModified(rec, time.Date(2026, 11, 7, 9, 30, 5, 0, time.FixedZone("CET", 3600)))

	want := "Sat, 07 Nov 2026 08:30:05 GMT"
	if got := rec.Header().Get("Last-Modified"); got != want {
		t.Fatalf("Last-Modified = %q, want %q (UTC IMF-fixdate)", got, want)
	}
}
