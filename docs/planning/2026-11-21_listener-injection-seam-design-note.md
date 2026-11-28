# Design note: the listener-injection seam (`ServiceConfig.Listener`)

**Date:** 2026-11-21 · Status: EXECUTED same session (core [Unreleased], rides the
next core tag) · Demand: FIRED twice — go-compliance ("appkit core binds TCP only,
`service.go:133-134`") and nsfw-classifier ("grep WithListener: zero hits") name
the SAME missing seam in their adoption reviews. Plan: SUPERB v17 leg (b),
`docs/planning/2026-11-21_01_56-SUPERB-v17-the-default-onramp-plan.html` T5/T6.

## Problem

`Service.Start` unconditionally binds TCP via `net.ListenConfig.Listen(ctx, "tcp",
cfg.Addr)`. A consumer that needs a unix-socket bind (local-only IPC, hub-style
daemon fronting), systemd socket activation (`sd_listen_fds` handoff), or a
pre-bound listener from any other source cannot host on appkit core — the flip
condition both waiting consumers recorded.

## Design of record

One config field, zero new types:

```go
// skip-validate
// ServiceConfig
Listener net.Listener // when set, used as-is instead of binding Addr
```

`Start` binds only when `Listener == nil`; otherwise the supplied listener is
used verbatim and `Addr` is ignored (the delegated server config already takes
its address from `listener.Addr().String()`).

### Contract pins

1. **Ownership transfers at `Start`.** From the moment `Start` is called appkit
   owns the listener: every synchronous failure path closes it (hook failure and
   `server.config_invalid` already close; documented for supplied listeners
   too), and graceful shutdown closes it through the delegated server — the
   identical lifecycle a self-bound listener gets. No "borrowed listener" mode.
2. **Single-serve ownership carries over.** A consumed listener cannot serve
   again (fresh `httputil.Server` per `Start`, upstream contract). A
   restart-after-shutdown flow must supply a FRESH listener per `Start`; a
   reused one fails with the listener-closed error class. Documented, pinned.
3. **`Addr()`/`Running()` are unchanged.** They read `s.ln`, which is the
   supplied listener after `Start`; `Addr()` reports the listener's own address
   (a unix socket's `Addr().String()` is its path). Post-listen `StartHooks`
   observe it (pinned by the existing contract tests through the new path).
4. **Classification is unchanged.** `server.listen_failed` still names only
   real bind failures (a supplied listener cannot produce one); the delegated
   rejection still surfaces as `server.config_invalid` with the supplied
   listener closed again; the six shutdown phase lines are untouched.
5. **Unix sockets are the canonical case, not a special mode.** The consumer
   binds `net.Listen("unix", path)` and passes it; appkit never learns the
   network. Client-side dialing stays the consumer's job (a stdlib
   `http.Transport` with a unix `DialContext`), matching the hub-daemon shape
   nsfw-classifier described.
6. **Timeouts and sentinels apply unchanged.** `ReadTimeout`/`WriteTimeout`/
   `NoTimeout` map onto the delegated server the same way — they are
   connection semantics, not bind semantics.
7. **TLS interplay (G1).** `ServerConfig.TLSConfig` stays nil on this seam;
   when Core TLS ships it composes with a supplied listener exactly as the
   spike §6 sketch describes (`Serve(tls.NewListener(ln, tlsCfg))`-shaped wrap
   at the delegated seam). The seam does not pre-empt that decision.
8. **Validate stays permissive.** No new rejection: `Listener != nil` with a
   set `Addr` is documented as "Addr ignored", not an error — the zero-value
   `Addr` path already defaults it, and rejecting harmless states is not this
   field's job.

### Test matrix (all in `listener_test.go`, hermetic, `localhost:0` / tmp sockets)

- pre-bound TCP listener serves `/health`; `Addr()` reports its address
- unix-socket listener serves `/health` over a unix-dialing client
- StartHook failure closes the SUPPLIED listener (ownership pin 1)
- `server.config_invalid` closes the SUPPLIED listener (ownership pin 1)
- restart-after-shutdown with a reused listener fails loudly (pin 2)

Integration carries the E2E: a service on a unix socket through the PUBLISHED
tag, probed with a unix `http.Client` (pins the consumer shape both reviews
asked for).

## Rejected alternatives

- `WithListener(...)` functional option — the config struct is the API here;
  every other ServiceConfig surface is a field, and an option type would be
  the only one of its kind.
- String `Addr` extension (`unix:///path`) — hides the listener lifecycle
  behind a parser, grows a second bind path in `Start`, and still cannot
  express socket activation. The listener IS the seam.
- `ListenConfig` exposure — over-configures bind-time control (keepalive,
  multipath TCP) nobody asked for; revisit if a consumer names it.
