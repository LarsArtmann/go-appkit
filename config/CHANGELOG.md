# Changelog

All notable changes to this module are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the module follows
semantic versioning.

## [Unreleased]

### Changed

- `collectFields` restructured to guard-clause style with a `nestedWalk`
  helper — behavior-identical, but the function drops from cyclomatic
  complexity 14 to under the module's cyclop limit of 12. The complexity
  finding shipped red in v0.1.0 (CI fleet-lint, latent since the release
  commit; caught in the 2026-11-06 CI-red triage).

## [0.1.0] - 2026-11-06

### Added — initial module (battery F1, SUPERB v12 C-train)

- `Load[T]` / `Loader[T]`: env-first struct-tag binding with fixed precedence
  **flag > env > file > default**; empty env/flag values count as unset.
- Tag grammar: `env:"NAME"`, `env:"NAME,required"`, `default:"..."`, and
  `env:"PREFIX"` on struct fields for nested key prefixes; embedded structs
  bind as promoted fields; untagged fields are never touched.
- Supported scalars: string, bool, all int/uint widths, float32/64,
  time.Duration, []string (comma-split). Tagged anything-else fails with
  `unsupported_type` — no silent coercion (the koanf WeaklyTypedInput trap is
  encoded as a test).
- File layer: `.yaml`/`.yml` via go-faster/yaml and `.json` via stdlib, strict
  unknown-key checking, lowercased dotted key space (`nested.addr`).
- Classified failures: one go-error-family **Rejection** (`config.invalid`)
  wrapping typed `*FieldError`s (field path + key + kind), collectable via
  `FieldErrorsOf`.
- `Keys[T]` — every declared env key, sorted (the push-model
  discoverability answer).
- Options: `WithFile`, `WithLookup`, `WithValues` (flag layer),
  `WithValidate` (cross-field hook, runs last).
- koanf verdict: DECLINED — design note at
  `docs/planning/2026-11-06_02-30_appkit-config-design-note.md`.
