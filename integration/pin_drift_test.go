package integration_test

import (
	"os"
	"strings"
	"testing"
)

// The documented pin map: what this module MUST resolve for its E2E tests to
// mean what the docs say they mean. A red test here means a release train (or
// a hand edit) moved a pin without updating the documented contract — the
// 2026-09-17 drift class this guard exists for (SUPERB plan v4 C1).
//
// Scope: only contract-relevant pins are fixture-pinned. The go-appkit family
// entries enforce the LATEST-published charter (scripts/check-pin-drift.sh
// additionally asserts each equals the module's newest tag); the cross-repo
// entries (cqrs-htmx, go-sse, httputil) are the composition contracts named in
// AGENTS.md. Routine dependabot bumps of other deps deliberately do NOT touch
// this fixture.
//
// PIN PHILOSOPHY (USER GATE §g-1, decided 2026-09-17 by execution order:
// LATEST-published-only retained). This is the flip point if the philosophy
// ever changes: to ALSO mirror cqrs-htmx setup's older resolution, change the
// four go-appkit entries to setup's versions and re-point
// scripts/check-pin-drift.sh check 2 at setup/go.mod instead of the newest
// family tag. Nothing else in this file or the script needs to change.
var documentedPins = map[string]string{
	"github.com/larsartmann/go-appkit":                      "v0.5.1",
	"github.com/larsartmann/go-appkit/errorpages":           "v0.1.0",
	"github.com/larsartmann/go-appkit/flightrecorderhealth": "v0.1.3",
	"github.com/larsartmann/go-appkit/health":               "v0.1.2",
	"github.com/larsartmann/go-appkit/otel":                 "v0.1.1",
	"github.com/larsartmann/go-appkit/realtime":             "v0.1.1",
	"github.com/larsartmann/go-appkit/security":             "v0.1.0",

	"github.com/larsartmann/cqrs-htmx/v4":   "v4.9.0",
	"github.com/larsartmann/go-sse":         "v0.6.0",
	"github.com/larsartmann/go-sse/ssetest": "v0.3.0",
	"github.com/larsartmann/httputil":       "v1.2.0",
}

// documentedGoDirective pins the language version this module (and the rest of
// the repository, per AGENTS.md and go.work) is documented against. The
// accidental go 1.27.1 directive bump on 2026-09-17 broke every local
// workspace command; a deliberate toolchain bump must land here, in
// integration/go.mod, in go.work, and in AGENTS.md in the same change.
const documentedGoDirective = "1.26.7"

func TestGoModPinsMatchDocumentedPins(t *testing.T) {
	t.Parallel()

	direct, _, _ := parseGoMod(t)

	for mod, want := range documentedPins {
		got, ok := direct[mod]

		if !ok {
			t.Errorf("go.mod no longer requires %s directly — update documentedPins", mod)

			continue
		}
		if got != want {
			t.Errorf("pin drift: %s resolved to %s, documented pin is %s (release train without the pin/doc bump?)",
				mod, got, want)
		}
	}

	for mod := range direct {
		if !strings.HasPrefix(mod, "github.com/larsartmann/go-appkit") {
			continue
		}
		if _, documented := documentedPins[mod]; !documented {
			t.Errorf("go.mod requires undocumented family module %s — add it to documentedPins", mod)
		}
	}
}

func TestGoModHasNoFilesystemReplaceDirectives(t *testing.T) {
	t.Parallel()

	_, _, hasReplace := parseGoMod(t)

	if hasReplace {
		t.Error("go.mod carries a replace directive — NEVER tag a module whose go.mod has one " +
			"(working-tree replaces are for cross-repo debugging only and must be removed before tagging)")
	}
}

func TestGoModGoDirectiveMatchesDocumentedToolchain(t *testing.T) {
	t.Parallel()

	_, goDirective, _ := parseGoMod(t)

	if goDirective != documentedGoDirective {
		t.Errorf("go directive is %s, documented toolchain is %s — a toolchain bump must update "+
			"integration/go.mod, go.work, root go.mod, and AGENTS.md together", goDirective, documentedGoDirective)
	}
}

// parseGoMod reads this module's go.mod (dependency-free: the module pins
// published tags and must not grow a golang.org/x/mod dependency just to read
// its own manifest) and returns its direct requires, go directive, and whether
// any replace directive exists.
func parseGoMod(t *testing.T) (map[string]string, string, bool) {
	t.Helper()

	data, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}

	direct := make(map[string]string)
	goDirective := ""
	hasReplace := false
	inRequireBlock := false

	for line := range strings.SplitSeq(string(data), "\n") {
		l := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(l, "require ("):
			inRequireBlock = true
		case inRequireBlock && l == ")":
			inRequireBlock = false
		case strings.HasPrefix(l, "replace"):
			hasReplace = true
		case strings.HasPrefix(l, "go "):
			goDirective = strings.TrimSpace(strings.TrimPrefix(l, "go "))
		case inRequireBlock, strings.HasPrefix(l, "require "):
			fields := strings.Fields(l)

			if len(fields) >= 2 && !strings.Contains(l, "// indirect") {
				direct[fields[0]] = fields[1]
			}
		}
	}

	return direct, goDirective, hasReplace
}
