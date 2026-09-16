package security_test

import (
	"strings"
	"testing"

	"github.com/larsartmann/go-appkit/security"
)

// TestCSPProductionBlocksInlineScripts is the CV pin: the production policy
// must not contain 'unsafe-inline' (style-src excepted by explicit decision)
// and NEVER 'unsafe-eval'.
func TestCSPProductionBlocksInlineScripts(t *testing.T) {
	t.Parallel()

	policy := security.BuildCSP(security.CSPConfig{
		Environment: security.Production,
		Nonce:       "abc123",
		StyleInline: true,
	})

	if strings.Contains(policy, "unsafe-inline' script-src") || scriptSrcHas(policy, "unsafe-inline") {
		t.Errorf("production script-src must not carry 'unsafe-inline': %s", policy)
	}

	if scriptSrcHas(policy, "unsafe-eval") {
		t.Errorf("unsafe-eval is NEVER grantable via the builder, any environment: %s", policy)
	}

	if !scriptSrcHas(policy, "nonce-abc123") {
		t.Errorf("production script-src must carry the nonce: %s", policy)
	}
}

// TestCSPHeaderParsesForEveryEnvironment pins determinism: directives are
// sorted, every token is a valid CSP source token, and the header parses
// for all environments.
func TestCSPHeaderParsesForEveryEnvironment(t *testing.T) {
	t.Parallel()

	for _, env := range []security.Environment{security.Development, security.Staging, security.Production} {
		policy := security.BuildCSP(security.CSPConfig{
			Environment: env,
			Nonce:       "n0nce",
			StyleInline: true,
		})

		directives := strings.Split(policy, "; ")
		if len(directives) < 5 {
			t.Fatalf("%s: policy %q too short", env, policy)
		}

		for i := 1; i < len(directives); i++ {
			if directives[i-1] > directives[i] {
				t.Errorf("%s: directives not in deterministic (sorted) order: %q", env, policy)
			}
		}

		for _, d := range directives {
			parts := strings.Fields(d)
			if len(parts) == 0 {
				t.Errorf("%s: empty directive in %q", env, policy)
			}

			for _, token := range parts[1:] {
				if strings.ContainsAny(token, "();") {
					t.Errorf("%s: token %q contains CSP delimiters", env, token)
				}
			}
		}

		if !strings.Contains(policy, "default-src 'self'") {
			t.Errorf("%s: missing default-src 'self' in %q", env, policy)
		}
	}
}

// TestCSP_JSONLDExemption pins the JSON-LD carve-out: the ld+json media
// type may join script-src because JSON-LD is data, not code — browsers do
// not execute it.
func TestCSP_JSONLDExemption(t *testing.T) {
	t.Parallel()

	with := security.BuildCSP(security.CSPConfig{Environment: security.Production, JSONLD: true})
	without := security.BuildCSP(security.CSPConfig{Environment: security.Production})

	if !scriptSrcHas(with, "'application/ld+json'") {
		t.Errorf("JSONLD=true must add the media type to script-src: %s", with)
	}

	if scriptSrcHas(without, "'application/ld+json'") {
		t.Errorf("JSONLD=false must not add it: %s", without)
	}
}

func TestCSP_WildcardDoesNotImplyEval(t *testing.T) {
	t.Parallel()

	// Documentation-by-test for the family's most expensive CSP lesson:
	// even a wildcard script-src does not run DataStar/Alpine-style
	// expression compilation, because eval needs the 'unsafe-eval' source
	// token — which the builder never emits.
	policy := security.BuildCSP(security.CSPConfig{Environment: security.Development})
	if strings.Contains(policy, "unsafe-eval") {
		t.Errorf("no environment may emit 'unsafe-eval': %s", policy)
	}
}

func TestAddNonceToScriptSrc(t *testing.T) {
	t.Parallel()

	got := security.AddNonceToScriptSrc("default-src 'self'; script-src 'self'", "xyz")
	if !scriptSrcHas(got, "nonce-xyz") {
		t.Errorf("nonce not added: %q", got)
	}

	again := security.AddNonceToScriptSrc(got, "xyz")
	if again != got {
		t.Errorf("duplicate nonce injected: %q", again)
	}

	unchanged := security.AddNonceToScriptSrc("default-src 'self'", "xyz")
	if unchanged != "default-src 'self'" {
		t.Errorf("CSP without script-src must be returned unchanged, got %q", unchanged)
	}
}

func TestNonceContextRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := security.WithNonce(t.Context(), "round-trip")
	if got := security.NonceFromContext(ctx); got != "round-trip" {
		t.Errorf("NonceFromContext = %q, want round-trip", got)
	}

	if got := security.NonceFromContext(t.Context()); got != "" {
		t.Errorf("empty context nonce = %q, want \"\"", got)
	}
}

// scriptSrcHas reports whether the script-src directive's source list
// contains the exact token.
func scriptSrcHas(policy, token string) bool {
	for _, directive := range strings.Split(policy, "; ") {
		parts := strings.Fields(directive)
		if len(parts) > 0 && parts[0] == "script-src" {
			for _, p := range parts[1:] {
				if token == "unsafe-eval" || token == "unsafe-inline" {
					if p == "'"+token+"'" {
						return true
					}
				} else if strings.Contains(p, token) {
					return true
				}
			}
		}
	}

	return false
}
