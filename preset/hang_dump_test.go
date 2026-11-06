package preset_test

// TEMPORARY diagnostic harness for the intermittent 5s listener_close hang.
// DELETE after root-cause. Not part of the shipped suite.

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit/preset"
	"github.com/larsartmann/go-appkit/testkit"
)

func TestHangDump(t *testing.T) {
	for i := range 10 {
		runHangDumpIter(t, i)
	}
}

func runHangDumpIter(t *testing.T, i int) {
	t.Helper()

	svc, err := preset.New(testConfig())
	if err != nil {
		t.Fatalf("iter %d: New: %v", i, err)
	}

	errCh, err := svc.Start()
	if err != nil {
		t.Fatalf("iter %d: Start: %v", i, err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for !svc.Running() {
		if time.Now().After(deadline) {
			t.Fatalf("iter %d: service did not start", i)
		}

		time.Sleep(time.Millisecond)
	}

	base := "http://" + svc.Addr().String()

	for _, path := range []string{"/healthz", "/readyz", "/startupz", "/health"} {
		resp := testkit.HTTPGetContext(t, context.Background(), base+path)

		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}

	done := make(chan error, 1)
	start := time.Now()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		done <- svc.Shutdown(ctx)
	}()

	select {
	case err := <-done:
		fmt.Fprintf(os.Stderr, "iter %d: healthy shutdown in %v err=%v\n", i, time.Since(start), err)

		if err != nil {
			t.Errorf("iter %d: shutdown error: %v", i, err)
		}
	case <-time.After(2500 * time.Millisecond):
		buf := make([]byte, 1<<20) //nolint:makezero // filled in place by runtime.Stack, never appended

		n := runtime.Stack(buf, true)

		fmt.Fprintf(
			os.Stderr,
			"=== HANG DUMP iter %d (shutdown age %v) ===\n%s\n=== END DUMP ===\n",
			i,
			time.Since(start),
			buf[:n],
		)

		if err := <-done; err != nil {
			t.Logf("iter %d: post-hang shutdown error: %v", i, err)
		}
	}

	select {
	case err := <-errCh:
		_ = err
	default:
	}
}
