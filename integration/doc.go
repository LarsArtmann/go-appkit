// Package integration hosts cross-module and cross-repo end-to-end
// composition tests for go-appkit.
//
// The module exists so that E2E seams spanning multiple modules — or reaching
// into sibling repositories, such as cqrs-htmx's transport.JournalSSEStore →
// realtime replay bridge — can be tested without adding cross-module test
// dependencies to any library go.mod. It carries no production code and is
// never released; consumers never import it.
//
// # Pin contract
//
// go.mod pins PUBLISHED tags only: this module always tests exactly what a
// fresh consumer resolves from the proxy — the LATEST published tag of every
// go-appkit family module it requires (deliberately NOT cqrs-htmx setup's
// older consumer pin). The specific versions are documented in exactly one
// checked place: the documentedPins fixture in pin_drift_test.go, which
// fails when go.mod drifts from it. On top of that fixture,
// scripts/check-pin-drift.sh asserts every family pin equals that module's
// newest git tag, so a release train shipped without the integration pin
// bump fails before push.
//
// Release-train rule: bumping a family pin means go.mod, the documentedPins
// fixture, and the module CHANGELOG move in the same change. If the pin
// philosophy (LATEST-only vs mirroring setup's resolution) ever changes, the
// flip point is documented at the top of pin_drift_test.go.
package integration
