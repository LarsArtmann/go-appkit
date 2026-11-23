# Go-Directive Floor Policy — decision record (2026-11-23)

Owner request (2026-11-22): "I want go 1.27" — the repo should declare
`go 1.27`, not the patch-pinned `go 1.27.4` every module carried since the
floor trains. This record states what the request MEANS after measurement,
what enforces it, and the (unauthorized) path to full uniformity.

## Measured constraints (2026-11-23, all reproduced this session)

1. **`go mod tidy` is the authority for each module's directive.** It raises
   a module's `go` line to the highest `go` version declared anywhere in its
   dependency graph. Published appkit tags and go-cqrs-lite declare patches,
   so a literal `go 1.27` in every go.mod does not survive one tidy:
   `go mod tidy -diff` rewrites 11 of 15 go.mods. Measured tidy-stable
   values: root/config/flightrecorderhealth/realtime hold `go 1.27`;
   cqrs+docs `go 1.27.2` (go-cqrs-lite v4 modules declare 1.27.2); httpx
   `go 1.27.3`; errorpages/flightrecorder/health/integration/otel/preset/
   security/systemd `go 1.27.4`.
2. **Go forces go.work to the max member directive.** A lower workspace line
   fails every workspace command — `go build ./...` printed ten
   `module X listed in go.work file requires go >= 1.27.Y, but go.work lists
   go 1.27` errors (measured 2026-11-23, restored within the hour). The
   BuildFlow go-work-sync re-raise of go.work to `go 1.27.4` (daemon commit
   74e4d80, 2026-11-23 02:57) was therefore CORRECT, not drift. Do not
   hand-lower go.work.
3. **`GOTOOLCHAIN` and the `go` directive are different namespaces.**
   `GOTOOLCHAIN=go1.27` is rejected ("go1.27 is a language version but not a
   toolchain version"); scripts resolve the toolchain via
   `scripts/lib/toolchain.sh` (`APPKIT_TOOLCHAIN_PATCH=1.27.4` belt,
   `go_toolchain_for <directive>`), smoke-tested on both branches.

## The policy (as shipped)

- **Published floor = the root go.mod's `go 1.27`.** Consumers see the floor
  through the core module; satellites declare their tidy-stable minimums,
  and `go get` demands the highest directive in the resolved chain (today
  1.27.4 via health/integration-class modules) until the family re-tags.
- **Local floor = go.work, tooling-managed**, currently `go 1.27.4` by
  constraint 2.
- **Patched toolchain = build layer.** Every README/CI command pins
  `GOTOOLCHAIN=go1.27.4`; the CVE-per-train history (6617/6809/6859) rides
  the belt, not the directive. Consumers building with their own toolchain
  no longer inherit the CVE floor from the directive — that is the one
  regression the reset accepts, documented in the root README.
- **Guards** (all green 2026-11-23):
  - `scripts/check-go-directives.sh` — module BELOW root floor = FATAL;
    go.work BELOW max member = FATAL (the workspace-break class, caught at
    commit time); `toolchain`-directive ban retained. Floor source is the
    root go.mod (non-self-referential — the original go.work-derived floor
    could not catch the go.work drift it had just suffered).
  - `scripts/check-go-tidy-stable.sh` — `go mod tidy -diff` must be empty
    per module (fleet-wide at pre-tag + the CI `go-tidy-stable` job; staged
    go.mods only at pre-commit so the hook stays seconds).
  - Both wired into `scripts/pre-tag-checks.sh` and CI (2026-11-23).

## The gvac window stays OPEN

11 modules sit at directives ABOVE the machine default toolchain (1.27.2) —
exactly the `b04fc20` precondition. `respect_patch_floor` remains live;
`check-go-directives.sh` (pre-commit FATAL + CI) is the standing net.
Re-run it after every BuildFlow pass / floor train.

## Path to full uniformity (NOT authorized; recorded for the ask)

Every go.mod at literal `go 1.27` requires (a) the appkit family re-tagged
with directives at or below 1.27 — a multi-module release wave, core first,
then dependents in reverse-dependency order — and (b) go-cqrs-lite's v4
modules dropping their `go 1.27.2`. Until both happen, `go mod tidy` and
the workspace rule (constraint 1+2) revert any hand-uniformity. Consumer
impact of the current state: `go get github.com/larsartmann/go-appkit@latest`
resolves a chain whose max directive is 1.27.4.

## Integration pairing note (same train, measured)

go-health-dashboard v0.13.1+ is the templ v0.3.1100 line AND requires
go-health v0.10.0; appkit/health v0.3.2 predates go-health's
`WithVersion` stamp, and go-health v0.10.0's `DefaultRoutes()` mounts
`GET /version` with a fail-loud panic when unstamped. Integration (the
reference consumer) bumps the dashboard to v0.13.3 and mounts probe routes
with the version endpoint unset (`versionlessProbeRoutes` — the documented
consumer escape hatch). Unblock: health's next train adopts dashboard
v0.13.x + go-health v0.10.0 and stamps by default; the escape hatch and the
integration fixture comment drop then.
