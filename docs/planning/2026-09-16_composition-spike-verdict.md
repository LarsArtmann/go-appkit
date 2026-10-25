# Composition Spike Verdict: `Service` on `httputil.Server` — BLOCKED on upstream API

**Created:** 2026-09-16 · **Type:** spike result (C19 of the SUPERB plan v2) · **Verdict: DO NOT refactor today.**

The architecture-lever TODO proposed composing `Service` on `httputil.Server`
(its v1.1.x `ListenerAddr`/`StartTLS`/double-start guard removed the original
"no listener access" justification). This spike executed the read-and-prove
phase. **Result: the byte-identical swap is NOT achievable against the
current httputil API** — the refactor would change public semantics, which
the plan's own guard forbids (and the byte-identical-vs-evolve USER GATE is
unanswered).

## 1. Field-by-field coverage table (ServerConfig vs ServiceConfig)

| appkit `ServiceConfig`                | httputil `ServerConfig`             | Mapping                                               |
| ------------------------------------- | ----------------------------------- | ----------------------------------------------------- |
| `Addr`                                | `Addr`                              | identity                                              |
| `ReadTimeout` (+`NoTimeout` sentinel) | `ReadTimeout`                       | appkit's `serverTimeout()` maps -1 → 0 before handoff |
| `ReadHeaderTimeout`                   | `ReadHeaderTimeout`                 | identity                                              |
| `WriteTimeout` (+`NoTimeout`)         | `WriteTimeout`                      | same sentinel mapping                                 |
| `IdleTimeout`                         | `IdleTimeout`                       | identity                                              |
| `ShutdownTimeout`                     | `ShutdownTimeout`                   | identity                                              |
| — (no TLS config; G1 deferred)        | `TLSConfig` + `StartTLS(cert, key)` | the actual prize: the future Core TLS option          |

Score: 6/7 identity + the TLS seam appkit lacks. On paper, clean.

## 2. The three pins — why the swap fails

1. **Bind-error contract (BLOCKER).** appkit's public `Start() (<-chan error, error)`
   returns bind errors SYNCHRONOUSLY (`listen_failed` Rejection, classified).
   httputil's `Start() <-chan error` binds internally and reports failures
   ASYNC on the channel (verified `server.go:200-216`: `Listen` inside
   `Start`, error delivered via `errChan`). A delegating `Service.Start`
   either changes its signature (API break) or breaks the sync-error
   behavior callers can rely on today. Both violate byte-identical.

2. **`Addr()`/`Running()` semantics.** appkit returns `net.Addr` and
   documents `nil` before `Start`; `Running()` is "has a bound listener"
   (mutex-guarded `s.ln`). httputil exposes `Addr() string` +
   `ListenerAddr() (net.Addr, bool)`. Adapter code could bridge it, but the
   nil-before-Start contract and the drain-flip interplay (appkit's
   `Shutdown` nils the listener to make double-shutdown a no-op) would rest
   on httputil's own started-flag/atomic semantics — a second source of
   truth for lifecycle state (split brain).

3. **Shutdown phase-log sequence + error classification.** The six
   `shutdown phase complete` lines are a grep-able contract
   (`shutdownlog_test.go` pins them). The `listener_close` phase wraps
   failures as `server.shutdown_failed` (Infrastructure). httputil's
   `Shutdown` applies its own `ShutdownTimeout` and error wrapping, so
   delegating re-wraps errors (code drift) and risks double-applying
   timeouts. Preserving the exact log/error contract means re-implementing
   the sequence anyway — at which point httputil.Server adds no
   lifecycle value, only the TLS seam.

## 3. What WOULD unblock it (upstream ask)

httputil exposing listener injection — `NewServerListener(ln net.Listener,
cfg ServerConfig, handler http.Handler)` (or a `Server.ListenerAddr`-based
equivalent) — would let appkit keep its synchronous bind, its own lifecycle
state, and delegate ONLY the `http.Server` ownership (timeouts + graceful
shutdown + `StartTLS`). Until then the refactor is negative-value: the
plumbing appkit owns shrinks by ~30 lines while gaining two semantic risks.

## 4. Disposition

- Core TLS option (G1) stays deferred and is now gated on the same upstream
  ask (appkit can call `StartTLS` only after the listener/API seam exists).
- The AGENTS "recommended refactor" note is corrected: the composition is
  BLOCKED, not pending.
- The API-posture USER GATE (byte-identical vs evolve) becomes moot until
  the upstream API exists; re-open both together if a consumer demands TLS.

Spike source evidence: httputil `server.go` (`Start` at :200, `StartTLS` at
:249, `NewServer` at :162); appkit `service.go` (`Start` at :105,
`Shutdown` at :166, `Addr` at :284, `Running` at :296);
`shutdownlog_test.go` (the phase-log pin).
