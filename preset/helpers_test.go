package preset_test

import (
	"io"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/preset"
	"github.com/larsartmann/go-appkit/testkit"
)

// httpTimeout bounds every test HTTP client.
const httpTimeout = 15 * time.Second

// getURL GETs url (context-aware, testkit-delegated), reads and closes the
// body, and returns status and body — no *http.Response crosses the boundary,
// so nothing leaks and nothing needs closing at call sites.
func getURL(t *testing.T, url string) (int, string) {
	t.Helper()

	resp := testkit.HTTPGetContext(t, t.Context(), url)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read %s: %v", url, err)
	}

	return resp.StatusCode, string(body)
}

// testConfig returns a Config with the two production-default-hostile knobs
// neutralized the same way testkit.NewService does for plain ServiceConfig:
// an ephemeral loopback address and no drain wait.
func testConfig() preset.Config {
	return preset.Config{
		Addr: "127.0.0.1:0",
		Timeouts: preset.TimeoutsConfig{
			DrainDelay: appkit.NoDrainDelay,
		},
	}
}

// servePreset boots a preset-built service through the full chain and
// returns the base URL (testkit.Serve handles teardown).
func servePreset(t *testing.T, svc *appkit.Service) string {
	t.Helper()

	return testkit.Serve(t, svc).FullChainURL
}
