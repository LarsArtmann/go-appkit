// Package config loads application configuration from declarative struct tags:
// one Load call binds environment variables, optional YAML/JSON files, defaults,
// and flag results onto a plain struct, returning classified go-error-family
// failures that name the exact field path.
//
// It exists because the LarsArtmann fleet kept hand-rolling the same
// envOr/envBool/envDurationOr pattern (22 config packages, 25 ad-hoc
// os.Getenv parsers across 39 Go apps at the 2026-11-06 audit). This package
// is that pattern as a library — env-first struct tags, fixed precedence,
// nothing else. See README.md for the full contract and the design note at
// docs/planning/2026-11-06_02-30_appkit-config-design-note.md for the koanf
// decline verdict.
package config
