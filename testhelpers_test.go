package appkit

import (
	"context"
	"errors"
	"net"
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

func freePort(t *testing.T) string {
	t.Helper()

	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}

	addr := listener.Addr().String()

	_ = listener.Close()

	return addr
}

func waitForRunning(t *testing.T, svc *Service) {
	t.Helper()

	deadline := time.Now().Add(testTimeout)

	for time.Now().Before(deadline) {
		if svc.Running() {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatal("service did not start within timeout")
}

func httpGet(t *testing.T, ctx context.Context, url string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request %s: %v", url, err)
	}

	return resp
}
