# Spike: leaf decoupling — systemd without the core module

**Date:** 2026-11-21 · SUPERB v17 leg (d), task T8 · Verdict: a near-free fix
exists; routed to TODO P2 to ride the next systemd train. Demand: nextd
("appkit's unit of adoption is the _service_, not a component") and audiblesync
name the same wedge — sd_notify adoptable only by taking the whole core module.

## T8.1 — the couplings, mapped

`systemd` imports `github.com/larsartmann/go-appkit` (the ROOT module: server
lifecycle + middleware stack + charmbracelet/log + httputil — the whole
indirect graph) for exactly ONE type: `appkit.Hook` (3 references in
`systemd/hooks.go`). And `appkit.Hook` is a **type ALIAS**:

```go
type Hook = func(context.Context) error
```

Tests additionally construct real `appkit.Service`s (`hooks_test.go`) — that
is a legitimate test-only use and can stay a test-scoped dependency.

## T8.2 — options

**Option A (RECOMMENDED): local alias, appkit goes test-only.** Define
`type Hook = func(context.Context) error` inside systemd and use it in the
`Hooks` struct. Because BOTH types are aliases of the same underlying type,
every existing call site keeps compiling unchanged: `systemd.Hook` values
assign into `appkit.ServiceConfig.StartHooks []appkit.Hook` and back — Go
assignability rules make the alias swap invisible. Consumer dep graph
collapses from the full root module (charmbracelet/log, lipgloss, httputil,
the server chain) to `go-systemd/v22 + go-error-family`. Zero API break, zero
migration note needed beyond the CHANGELOG line.

**Option B: a shared hooks package.** A `go-appkit/hooks` (or family) module
owning the alias, imported by core and leaves alike. Solves the same problem
by adding a module — over-structured for ONE type alias that unifies anyway.
Declined: option A gets the identical outcome with one deleted require.

**Option C: keep the coupling.** The status quo — "component" adopters must
take core. This is exactly the wedge the two decliners named; keeping it
contradicts the default-onramp mission (SUPERB v17). Declined.

## T8.3 — routing

- TODO_LIST P2 row: "systemd leaf decoupling (T8 spike verdict: option A —
  local `Hook` alias, go-appkit require → test-only; consumers' dep graph
  drops the root module; aliases unify, zero API break)".
- Rides the next systemd tag train (go.mod require change + CHANGELOG;
  `hooks_test.go` keeps appkit as a test dep; the go.mod `require` block
  gains no new module — appkit moves from direct to test-scoped).
- The same alias pattern is the REPLAY for any future leaf module tempted to
  import root for types only (check `rg 'appkit\.' <module> --glob '!*_test.go'`
  before adopting root into a satellite's direct require).

## Acceptance (for the executing train)

`systemd` builds with `go-appkit` absent from its non-test import graph
(`go mod why -m github.com/larsartmann/go-appkit` reports test-only); the
module suite `-race` green unchanged; the go.mod diff shows the require
annotation flip and nothing else.
