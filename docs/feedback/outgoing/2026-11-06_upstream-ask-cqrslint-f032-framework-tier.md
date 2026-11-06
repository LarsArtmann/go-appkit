# Upstream draft: go-cqrs-lite cqrs-lint — F032's catalog mandate mis-fires on framework wrappers, and their consumers escape it entirely

**Status:** DRAFT — filing is USER-gated (rule verified against the v4.20.0 tag source AND
master at commit time; see Evidence). Do not file without explicit approval.
**Target repo:** github.com/LarsArtmann/go-cqrs-lite (own repo — design note, not an
external ask)
**Evidence versions:** cmd/cqrs-lint/v4 v4.20.0 (published latest; `pkg/rules/adoption/f032.go`
byte-identical on master at 2026-11-06); affected consumer: go-appkit `cqrs` module
(cqrs/v0.8.0 pin), suppressed inline 2026-11-06.

## The ask

F032 (`no-catalog-import`, owner policy 2026-11-03) fires on any project whose
production code imports a non-catalog go-cqrs-lite module. Two shapes break:

1. **Framework wrappers own no domain.** go-appkit's `cqrs` module wraps
   `system`/`command`/`query`/`decider` — a typed facade over the consumer's
   own `DomainConfig`. It has no commands, events, or queries of its own to
   declare; demanding `catalog.NewBuilder()` here means inventing vocabulary
   that belongs to whoever consumes the wrapper. We suppressed inline with
   that reason (established repo pattern), but a rule that every framework
   wrapper must hand-suppress is a rule without a tier model — E023 already
   carries one ("engine-tier modules are exempt by design").

2. **The mandate never reaches the apps that matter.** F032's detector keys on
   DIRECT go-cqrs-lite imports (`firstLibraryImport` scans `ctx.GoFiles`
   imports). An app composing through go-appkit/cqrs never imports
   go-cqrs-lite directly — the policy's actual audience (the app that owns
   the domain) gets no coaching, while the wrapper that owns none gets the
   error. The policy gap and the false positive are the same coin.

## Proposed shapes (either, or both)

- **Tier exemption:** F032 skips modules whose go.mod module path is imported
  BY other modules as a library (hard) or honors a `.cqrs-lint.json` opt-out
  key like `"framework-wrapper": true` with a recorded reason (easy, honest).
- **Reach-through coaching:** when the analyzed project has no direct
  go-cqrs-lite import but its require graph contains one via a wrapper module,
  emit an INFO (not an error) naming the wrapper and suggesting the
  systemcatalog derivation at the app boundary — that is where the policy
  can actually be satisfied.

## What we did locally (no dependency on the ask)

- `cqrs/commands.go`: `// cqrs-lint:ignore(F032) framework wrapper, not a
  domain app: consumers derive their catalog from their own DomainConfig
  (systemcatalog; F032 policy 2026-11-03)` at the anchored import, plus the
  rationale in the module CHANGELOG [Unreleased].
- CI note: the cqrs-lint job installs `@latest`, so new ERROR-severity rules
  retro-red every consumer CI until suppressed — pinning the install (the
  art-dupl-gate precedent) is queued on go-appkit's side regardless of this
  ask's outcome.

## Verification trail

- Rule source read from the v4.20.0 proxy zip (`module_catalog.go`, `f032.go`,
  `register_consumer.go`): unconditional registration in the consumer set,
  no tier/profile gate before or on master.
- Local repro: `cd cqrs && GOWORK=off cqrs-lint .` → `ERROR commands.go:13
  [F032]` before the suppression, exit 0 after; the same binary family the
  CI `@latest` install resolved to (v4.20.0, proxy `@latest` verified
  2026-11-06).

💘 Generated with Crush
