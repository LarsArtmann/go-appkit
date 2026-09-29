# Upstream draft: go-licenses — publish builds that track a current Go toolchain (load go.mod floors ≥ 1.27)

**Status:** DRAFT — filing is USER-gated (T15 batch). Re-verify the
reproduction below on the target machine before filing; the evidence was
recorded locally during go-appkit's BuildFlow configuration on
2026-09-29 and the exact embedded-toolchain version should be quoted
from a fresh run.
**Target repo:** github.com/google/go-licenses
**Evidence context:** go-appkit pins `go 1.27.1` in every go.mod
(repo-wide toolchain unification, 2026-09-29); license checking is wired
through `nix run nixpkgs#go-licenses` inside BuildFlow and had to be
skipped for this repo.

## The ask

`go-licenses` binaries distributed today (including the nixpkgs package,
which builds the released version with an older Go) fail to LOAD modules
whose go.mod declares a recent `go` directive — the embedded toolchain
rejects the module before any license checking happens, with a
`go.mod requires go >= 1.27.1` class of error. Consumers on current Go
floors therefore cannot use the tool at all; there is no flag to loosen
the load check because the failure is in module loading, not in the
license graph.

Ask: publish (and have distro packages build with) a toolchain that
tracks supported Go releases — or, more durably, adopt the same
toolchain-switching behavior `go test`/`go build` have (GOTOOLCHAIN
awareness), so the binary can load modules with newer directives
instead of failing hard.

## Evidence

Reproduction (run in the go-appkit repo, go 1.27.1 floor):

```bash
nix run nixpkgs#go-licenses -- check ./...
# fails at module load: go.mod requires go >= 1.27.1 (embedded toolchain is older)
```

Downstream impact recorded in this repo's build config: `.buildflow.yml`
sets `skip: license-check` with the comment that the nix-run go-licenses
embeds a pre-1.27 go and cannot load the repo — meaning the repo's
entire dependency tree ships without license verification until the
tool can load it.

## Why upstream

This is not a nix packaging bug alone: any consumer with a
current-release go directive hits the same wall with distro-packaged
binaries, and go-licenses' value collapses exactly for projects that
track new Go releases promptly. Toolchain awareness would fix the whole
class.
