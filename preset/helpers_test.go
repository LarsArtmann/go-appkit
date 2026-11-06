package preset_test

import (
	"testing"

	"github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/preset"
)

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
