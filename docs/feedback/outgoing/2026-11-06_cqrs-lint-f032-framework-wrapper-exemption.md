# cqrs-lint F032: framework-wrapper exemption proposal

**To:** go-cqrs-lite maintainers
**From:** go-appkit (framework consumer, cqrs module owner)
**Date:** 2026-11-06
**Status:** local position recorded + inline suppression shipped; filing upstream when the go-cqrs-lite issue tracker window opens (owner-gated: the fleet is GOPRIVATE and issue filing is a maintainer act)

## The rule

F032 (catalog-mandate) requires every module that imports go-cqrs-lite to
derive and wire a domain catalog (`systemcatalog`), enforcing that CQRS
consumers actually model a domain rather than bolting on command buses ad
hoc.

## Why framework wrappers cannot comply

`github.com/larsartmann/go-appkit/cqrs` is a framework wrapper, not a
domain app: it composes go-cqrs-lite's `system` engine into appkit's
service lifecycle and exposes typed command/query facades. Its consumers
never import go-cqrs-lite directly — they derive catalogs from THEIR OWN
`DomainConfig` through `systemcatalog` at composition time. There is no
domain vocabulary at the wrapper layer to put in a catalog, and inventing
one would be a fabrication: the wrapper's mandate is to be invisible.

The mandate therefore cannot reach either side of the wrapper:

- the wrapper has no domain to catalog (F032 fires and can only be
  suppressed);
- the consumer satisfies F032 through the wrapper's composition anyway
  (the catalog exists at THEIR composition root), so a blanket exemption
  does not weaken the rule's real target.

## The convention we shipped (what this memo records)

Suppression inline at the `command/v4` import, reason-bearing, module
CHANGELOG-recorded (see `cqrs/CHANGELOG.md` [Unreleased], 2026-11-06):

> this module is a framework wrapper over `system`, not a domain app; the
> domain vocabulary belongs to consumers, who derive catalogs from their
> own DomainConfig via `systemcatalog`.

## The ask

1. A first-class wrapper marker so F032 (and future rules in the same
   family) can distinguish framework wrappers from domain apps without a
   suppression — candidates: a build-tag, a doc-comment convention
   (`// cqrs:framework-wrapper` at the package clause), or a config
   allowlist keyed by module path.
2. Until then, bless the reason-bearing inline suppression as the official
   pattern (the linter already warns on misplaced suppressions; anchoring
   the F032 suppression at the import it exempts keeps that check
   meaningful).

## Verification status

- cqrs-lint v4.21.0 green on the cqrs tree with the suppression
  (verified 2026-11-06: `GOWORK=off go run
  github.com/larsartmann/go-cqrs-lite/cmd/cqrs-lint/v4@v4.21.0 .` exit 0).
- CI installs cqrs-lint PINNED at v4.21.0 as of the same day (the `@latest`
  install made CI retro-red twice; the art-dupl-gate pin precedent applies).
