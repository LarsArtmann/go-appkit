package preset_test

import (
	"testing"
	"time"

	"github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/preset"
	"github.com/larsartmann/go-appkit/testkit"
)

// httpTimeout bounds every test HTTP client.
const httpTimeout = 15 * time.Second

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
