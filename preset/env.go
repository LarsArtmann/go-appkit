package preset

import (
	"fmt"
	"os"

	errorfamily "github.com/larsartmann/go-error-family"
)

// Frozen env-switch names (design note §2 — do not add per-app switches).
const (
	envOTEL       = "APPKIT_OTEL"
	envRealtime   = "APPKIT_REALTIME"
	envErrorPages = "APPKIT_ERRORPAGES"
	envSecurity   = "APPKIT_SECURITY"
)

// triBool parses an on|off env var. Absent (or empty) keeps absentVal;
// anything else is a Rejection naming the variable and the valid values —
// boot fails before any listener binds.
func triBool(name string, absentVal bool) (bool, error) {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		return absentVal, nil
	}

	switch value {
	case "on":
		return true, nil
	case "off":
		return false, nil
	default:
		return false, errorfamily.NewRejection(
			"preset.env_invalid", fmt.Sprintf("invalid %s=%q: want on|off", name, value),
		)
	}
}

// parseSecurityProfile parses the none|standard|strict enum. Absent (or
// empty) means none — security never silently joins a stack.
func parseSecurityProfile() (SecurityProfile, error) {
	value, ok := os.LookupEnv(envSecurity)
	if !ok || value == "" {
		return SecurityNone, nil
	}

	switch SecurityProfile(value) {
	case SecurityNone, SecurityStandard, SecurityStrict:
		return SecurityProfile(value), nil
	default:
		return "", errorfamily.NewRejection(
			"preset.env_invalid", fmt.Sprintf("invalid %s=%q: want none|standard|strict", envSecurity, value),
		)
	}
}

// resolvedLegs is the post-precedence decision for every leg: explicit code
// (non-nil Config field) wins over env in both directions.
type resolvedLegs struct {
	otel        bool
	realtime    bool
	errorPages  bool
	security    SecurityProfile
	hasSecurity bool
}

// resolveLegs applies the grammar: nil leg field -> the env switch decides;
// non-nil leg field -> the leg's Enabled field decides and the env var is
// ignored for that leg.
func resolveLegs(cfg Config) (resolvedLegs, error) {
	legs := resolvedLegs{} //nolint:exhaustruct_v5 // zero-value start — every field is assigned below

	var err error
	if legs.otel, err = triBool(envOTEL, false); err != nil {
		return legs, err
	}

	if cfg.OTEL != nil {
		legs.otel = cfg.OTEL.Enabled
	}

	if legs.realtime, err = triBool(envRealtime, false); err != nil {
		return legs, err
	}

	if cfg.Realtime != nil {
		legs.realtime = cfg.Realtime.Enabled
	}

	if legs.errorPages, err = triBool(envErrorPages, true); err != nil {
		return legs, err
	}

	if cfg.ErrorPages != nil {
		legs.errorPages = cfg.ErrorPages.Enabled
	}

	if legs.security, err = parseSecurityProfile(); err != nil {
		return legs, err
	}

	legs.hasSecurity = legs.security != SecurityNone

	if cfg.Security != nil {
		profile := cfg.Security.Profile
		if profile == "" {
			profile = SecurityStandard
		}

		legs.security = profile
		legs.hasSecurity = cfg.Security.Enabled && profile != SecurityNone
	}

	return legs, nil
}
