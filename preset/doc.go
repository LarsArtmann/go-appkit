// Package preset composes the go-appkit family into a one-call service
// bootstrap: the always-on spine (core Service, sd_notify, health dashboard,
// error pages, error classification, version stamp) plus opt-in legs
// (OpenTelemetry, security profiles, realtime SSE) driven by four frozen
// environment switches with explicit-code-wins precedence.
//
// Grammar (frozen — see docs/planning/2026-11-06_appkit-preset-design-note.md):
//
//	APPKIT_OTEL        on|off      absent = off
//	APPKIT_REALTIME    on|off      absent = off
//	APPKIT_ERRORPAGES  on|off      absent = on (spine member)
//	APPKIT_SECURITY    none|standard|strict   absent = none
//
// A nil leg field in Config means the env switch decides; a non-nil leg
// decides itself via its Enabled field and the env var is ignored. Invalid
// values fail fast at New with a Rejection naming the variable and the valid
// values — before any listener binds.
//
// The preset composes shipped batteries only; it never invents credentials
// (Metrics passthrough), never wraps Run (the consumer owns blocking and
// route registration), and never lets security silently join a stack.
package preset
