package appkit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testTimeout = 2 * time.Second

func newHealthRequest(t *testing.T) *http.Request {
	t.Helper()

	return httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
}

func httptestNewRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()

	if rec.Code != want {
		t.Errorf("status = %d, want %d", rec.Code, want)
	}
}

func assertHealthBody(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()

	body := rec.Body.String()
	if body != want+"\n" {
		t.Errorf("body = %q, want %q", body, want+"\n")
	}
}

func assertShutdownCalled(t *testing.T, called <-chan struct{}, msg string) {
	t.Helper()

	select {
	case <-called:
	case <-time.After(testTimeout):
		t.Fatal(msg)
	}
}

func assertServerStopped(t *testing.T, errCh <-chan error) {
	t.Helper()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("server returned error: %v", err)
		}
	case <-time.After(testTimeout):
		t.Fatal("server did not stop")
	}
}

func assertErrorWithin(t *testing.T, errCh <-chan error, wantErr error) {
	t.Helper()

	select {
	case err := <-errCh:
		if !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	case <-time.After(testTimeout):
		t.Fatal("error was not received within timeout")
	}
}

func newMockShutdown(called chan struct{}) func(context.Context) error {
	return func(_ context.Context) error {
		close(called)

		return nil
	}
}

func expectError(t *testing.T, err error, msg string) {
	t.Helper()

	if err == nil {
		t.Fatal(msg)
	}
}

func expectOpenSQLiteError(t *testing.T, path, msg string) {
	t.Helper()

	_, err := OpenSQLite(context.Background(), SQLiteConfig{Path: path})
	expectError(t, err, msg)
}
