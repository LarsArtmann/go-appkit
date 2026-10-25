# Upstream draft: httputil — v1.6.0 was tagged with a filesystem `replace` in its go.mod

**Status:** DRAFT — filing is USER-gated (source-verified against the proxy-served tag; see Evidence). Do not file without explicit approval.
**Target repo:** github.com/larsartmann/httputil (own repo — decision note, not an external ask)
**Evidence versions:** httputil v1.6.0 (published latest, tagged 2026-10-23 window; pin-drift flagged it 2026-10-24).

## The ask

The v1.6.0 tag's `go.mod` (as served by proxy.golang.org) contains:

```
replace github.com/larsartmann/httputil/server_timing => ./server_timing
```

This violates the family's own release rule — go-appkit's pre-tag-checks
assert every tagged artifact carries NO filesystem `replace`, exactly
because a dependency's replace directives are not honored in module mode
and tooling treats the module as non-canonical:

- `go doc github.com/larsartmann/httputil@v1.6.0 <symbol>` refuses to load
  the module: "contains one or more replace directives. It must not
  contain directives that would cause it to be interpreted differently
  than if it were the main module."
- Consumers are unaffected at build time (verified: a scratch module
  requiring v1.6.0 tidies and builds clean — the directive is ignored in
  dependency position), but anything that loads the module AS the main
  module (go doc, pkgsite-style rendering, gopls module off-side tooling)
  errors.

If the replace is leftover from the server_timing extraction (the
subpackage now ships its own go.mod inside the zip, making the directive
a self-referential no-op), the fix is to delete the line and tag v1.6.1.

## Evidence

1. `$GOMODCACHE/cache/download/github.com/larsartmann/httputil/@v/v1.6.0.mod`
   (proxy-served, sumdb-verified) — line 15: `replace
   github.com/larsartmann/httputil/server_timing => ./server_timing`.
2. `GOWORK=off go doc github.com/larsartmann/httputil@v1.6.0 Language` →
   the replace-directive refusal quoted above (2026-10-25).
3. Scratch consumer (`go 1.27.2`, requires httputil v1.6.0, imports the
   root package): `go mod tidy` + `go build ./...` both exit 0 — the
   failure mode is tooling-side, not build-breaking.

## Disposition

- v1.6.1 (or a re-tag per the family's force-move ritual, with the
  go.sum-cache repair recipe from AGENTS "Stale go VCS cache") — upstream's
  call.
- go-appkit side: integration re-pinned to v1.6.0 in-tree (pin-drift green
  2026-10-25); the bump rides the next natural integration train. otel
  rode to v1.6.0 the same day for the LoggingFromContext recipe.
