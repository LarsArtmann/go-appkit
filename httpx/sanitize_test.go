package httpx_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/go-appkit/httpx"
	errorfamily "github.com/larsartmann/go-error-family"
)

// leakingInfra is a raw server-side error carrying every standing leak
// class at once: filesystem paths, a source location, SQL, and a
// goroutine header. The no-leak contract says its TEXT never reaches the
// wire: server-side families render the family default message.
func leakingDetail() error {
	return errors.New("open /home/lars/secret/creds.json: C:\\Users\\admin\\keys.pem query " +
		"SELECT name FROM users WHERE password='hunter2' failed (goroutine 1, sanitize.go:59)")
}

func TestSanitize_ServerFamiliesNeverEchoRawText(t *testing.T) {
	t.Parallel()

	for _, family := range []errorfamily.Family{
		errorfamily.Transient,
		errorfamily.Corruption,
		errorfamily.Infrastructure,
	} {
		err := errorfamily.WrapInfrastructure(leakingDetail(), "db.creds_unreadable", "credentials unavailable")

		got := httpx.SanitizeErrorMessage(err)

		for _, leak := range []string{
			"/home/lars", "C:\\Users", "SELECT * FROM users", "hunter2", "goroutine 1", "sanitize.go:59",
		} {
			if strings.Contains(got, leak) {
				t.Errorf("family %v: sanitized message %q leaks %q", family, got, leak)
			}
		}
	}
}

// An UNCLASSIFIED error with leaking detail also renders the family
// default: unclassifiable defaults to a server-side family, and raw text
// from unknown sources is the least trustworthy of all.
func TestSanitize_UnclassifiedLeakDefaulted(t *testing.T) {
	t.Parallel()

	got := httpx.SanitizeErrorMessage(leakingDetail())

	for _, leak := range []string{"/home/lars", "SELECT", "goroutine"} {
		if strings.Contains(got, leak) {
			t.Errorf("sanitized message %q leaks %q", got, leak)
		}
	}
}

// Client-facing families keep the error's own text (it IS the API
// response) but leak patterns are scrubbed from it.
func TestSanitize_ClientFacingTextScrubbed(t *testing.T) {
	t.Parallel()

	err := errorfamily.Newf(errorfamily.Rejection, "config.bad",
		"cannot read /etc/appkit/config.yaml for user app")

	got := httpx.SanitizeErrorMessage(err)

	if strings.Contains(got, "/etc/appkit") {
		t.Errorf("scrubbed message %q leaks the path", got)
	}

	// A path is redacted whole (filename included); the surrounding
	// user-facing words survive so the message stays useful.
	if !strings.Contains(got, "cannot read") || !strings.Contains(got, "[redacted]") {
		t.Errorf("scrubbed message %q lost the surrounding words or the redaction marker", got)
	}
}

// The explicit opt-out wins, mirroring errorpages' Public() ladder:
// vouched-for text reaches the client verbatim. The Message() method
// classified errors carry is deliberately NOT honored (it would let
// authored text bypass the scrub; Public() is the deliberate-exception
// hatch).
func TestSanitize_ExplicitOptOuts(t *testing.T) {
	t.Parallel()

	vouched := publicMessageError{
		err:    leakingDetail(),
		public: "the export is not ready yet",
	}

	if got := httpx.SanitizeErrorMessage(vouched); got != "the export is not ready yet" {
		t.Errorf("Public() opt-out = %q, want the vouched text verbatim", got)
	}

	// The same raw error without the vouch defaults by family: the
	// Message() method on classified errors must not smuggle text past
	// the gate.
	cleaned := cleanMessageError{err: leakingDetail(), message: "export queued"}

	if got := httpx.SanitizeErrorMessage(cleaned); got == "export queued" {
		t.Error("Message() must not bypass the family gate")
	}
}

func (c cleanMessageError) Error() string { return c.err.Error() }

func (c cleanMessageError) Message() string { return c.message }

func (p publicMessageError) Error() string { return p.err.Error() }

func (p publicMessageError) Public() string { return p.public }

// End to end: the error body on the wire carries the sanitized message
// and never the raw internals.
func TestWriteError_NoLeakOnTheWire(t *testing.T) {
	t.Parallel()

	handler := httpx.ResultHandler(func(_ *httpx.Context) (*thing, error) {
		return nil, errorfamily.WrapInfrastructure(leakingDetail(), "db.creds_unreadable", "credentials unavailable")
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/things", nil))

	body := rec.Body.String()

	for _, leak := range []string{"/home/lars", "hunter2", "SELECT", "goroutine", ".go:"} {
		if strings.Contains(body, leak) {
			t.Errorf("wire body %q leaks %q", body, leak)
		}
	}

	var resp struct {
		Family  string `json:"family"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode wire body: %v", err)
	}

	if resp.Family != "infrastructure" {
		t.Errorf("family = %q, want infrastructure", resp.Family)
	}
}

type cleanMessageError struct {
	err     error
	message string
}

type publicMessageError struct {
	err    error
	public string
}
