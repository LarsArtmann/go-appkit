package preset

import (
	"strings"
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
)

// The grammar unit tests pin resolveLegs deterministically: defaults,
// explicit-code-wins (both directions), and invalid values failing as a
// Rejection that names the variable. The full boot-level matrix (what the
// legs DO) lives in the *_test.go integration files.

func TestResolveLegs_AllAbsent(t *testing.T) {

	legs, err := resolveLegs(Config{})
	if err != nil {
		t.Fatalf("resolveLegs: %v", err)
	}

	if legs.otel {
		t.Error("otel on with APPKIT_OTEL absent, want off")
	}

	if legs.realtime {
		t.Error("realtime on with APPKIT_REALTIME absent, want off")
	}

	if !legs.errorPages {
		t.Error("errorPages off with APPKIT_ERRORPAGES absent, want on (spine member)")
	}

	if legs.hasSecurity || legs.security != SecurityNone {
		t.Errorf("security = %q (has %t) with APPKIT_SECURITY absent, want none", legs.security, legs.hasSecurity)
	}
}

func TestResolveLegs_CodeWinsBothDirections(t *testing.T) {
	t.Setenv("APPKIT_OTEL", "on")
	t.Setenv("APPKIT_REALTIME", "on")
	t.Setenv("APPKIT_ERRORPAGES", "off")
	t.Setenv("APPKIT_SECURITY", "strict")

	legs, err := resolveLegs(Config{
		OTEL:       &OTELConfig{Enabled: false},
		Realtime:   &RealtimeConfig{Enabled: false},
		ErrorPages: &ErrorPagesConfig{Enabled: true},
		Security:   &SecurityConfig{Enabled: false},
	})
	if err != nil {
		t.Fatalf("resolveLegs: %v", err)
	}

	if legs.otel {
		t.Error("code-disabled otel overridden by env")
	}

	if legs.realtime {
		t.Error("code-disabled realtime overridden by env")
	}

	if !legs.errorPages {
		t.Error("code-enabled errorPages overridden by env")
	}

	if legs.hasSecurity {
		t.Error("code-disabled security overridden by env")
	}
}

func TestResolveLegs_EnvDecidesWhenCodeNil(t *testing.T) {
	t.Setenv("APPKIT_OTEL", "on")
	t.Setenv("APPKIT_REALTIME", "on")
	t.Setenv("APPKIT_ERRORPAGES", "off")
	t.Setenv("APPKIT_SECURITY", "standard")

	legs, err := resolveLegs(Config{})
	if err != nil {
		t.Fatalf("resolveLegs: %v", err)
	}

	if !legs.otel || !legs.realtime || legs.errorPages {
		t.Errorf("tri-bool legs = otel:%t realtime:%t errorPages:%t, want on/on/off", legs.otel, legs.realtime, legs.errorPages)
	}

	if legs.security != SecurityStandard || !legs.hasSecurity {
		t.Errorf("security = %q (has %t), want standard", legs.security, legs.hasSecurity)
	}
}

func TestResolveLegs_EmptyProfileDefaultsToStandard(t *testing.T) {

	legs, err := resolveLegs(Config{Security: &SecurityConfig{Enabled: true}})
	if err != nil {
		t.Fatalf("resolveLegs: %v", err)
	}

	if legs.security != SecurityStandard {
		t.Errorf("security = %q, want standard (empty profile defaults)", legs.security)
	}
}

func TestResolveLegs_InvalidValuesReject(t *testing.T) {
	for _, tc := range []struct {
		name string
		env  string
		want string
	}{
		{name: "otel", env: "APPKIT_OTEL", want: "APPKIT_OTEL"},
		{name: "realtime", env: "APPKIT_REALTIME", want: "APPKIT_REALTIME"},
		{name: "errorpages", env: "APPKIT_ERRORPAGES", want: "APPKIT_ERRORPAGES"},
		{name: "security", env: "APPKIT_SECURITY", want: "APPKIT_SECURITY"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(tc.env, "banana")

			_, err := resolveLegs(Config{})
			if err == nil {
				t.Fatalf("%s=banana accepted, want Rejection", tc.env)
			}

			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not name %s", err, tc.want)
			}

			if errorfamily.Classify(err) != errorfamily.Rejection {
				t.Errorf("error %v classifies as %v, want Rejection", err, errorfamily.Classify(err))
			}
		})
	}
}

func TestTriBool_CaseSensitiveNoTrim(t *testing.T) {
	t.Setenv("APPKIT_OTEL", " ON ")

	if got, err := triBool("APPKIT_OTEL", false); err == nil {
		t.Errorf(`" ON " accepted as %t — values are case-sensitive, no trimming`, got)
	}
}
