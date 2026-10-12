package integration

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestMain enforces this module's hermetic execution contract: the package
// must never compile or run under a go.work workspace (README "Build &
// verify"; pin charter in doc.go). A workspace's LOCAL members replace the
// PUBLISHED-tag pins this module exists to prove — the charter is then
// silently violated (tests still pass against local code, which is exactly
// the failure mode this guard makes loud).
//
// Covered paths: a future charter leak that re-adds ./integration to go.work
// would make root-launched `go test ./integration/...` resolve cleanly
// against local code — this check turns that into a hard failure. Bare
// invocations inside integration/ already fail at setup ("directory prefix .
// does not contain modules listed in go.work"). Standalone module mode (no
// go.work found — e.g. the directory opened on its own) resolves via the
// proxy and is hermetic-equivalent: it passes.
func TestMain(m *testing.M) {
	cmd := exec.Command("go", "env", "GOWORK")
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"hermetic guard: cannot probe `go env GOWORK` (%v) — refusing to run unverified\n", err)
		os.Exit(2)
	}
	resolved := strings.TrimSpace(string(out))
	if resolved != "" && resolved != "off" {
		fmt.Fprintf(os.Stderr,
			"hermetic violation: build governed by workspace %q — the integration module must run WITHOUT a go.work\n"+
				"(pin charter, doc.go): workspace members resolve the family pins to LOCAL code instead of\n"+
				"the published tags consumers actually get. If ./integration leaked into go.work again, run\n"+
				"scripts/check-workspace-charter.sh --fix . Otherwise verify hermetically:\n"+
				"  cd integration && GOWORK=off GOTOOLCHAIN=go1.27.2 go test ./... -race -count=1\n", resolved)
		os.Exit(2)
	}
	os.Exit(m.Run())
}
