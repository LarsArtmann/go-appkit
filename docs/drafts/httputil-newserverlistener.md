# Issue: expose listener injection (`NewServerListener`) so wrappers can own the bind

**Target:** github.com/larsartmann/httputil
**Kind:** feature request
**Verified:** 2026-10-13 against master `332abed3` (NewServer at server.go:162; Start binds internally at :221 via `listenConfig.Listen`; the `listener` field is private and has no injection path)

---

## Problem

`Server.Start` binds its own listener and delivers bind errors
asynchronously, which is the right default for standalone use — but any
wrapper that needs to own the bind (synchronous `listen_failed`
classification, its own readiness/lifecycle state, pre-serve startup
hooks that observe `Addr()`) cannot compose on top of `Server` today.
The listener is a private field with no injection seam.

## Ask

Expose listener injection, e.g.:

```go
// NewServerListener creates a Server serving on an already-bound listener.
// Ownership semantics match NewServer; Start/Shutdown operate on the
// injected listener instead of binding.
func NewServerListener(ln net.Listener, cfg ServerConfig, handler http.Handler) (*Server, error)
```

(or a `ServerConfig.Listener` field / `Server.ServeListener` method —
the shape is yours; the essential property is: caller-bound listener +
your `http.Server` ownership: timeouts, graceful shutdown, `StartTLS`).

## Consumer evidence

go-appkit's `appkit.Service` pinned a synchronous-start contract
(`listen_failed` classified at `Start()` return, `Addr()`/`Running()`
observable inside post-listen startup hooks, drain-window semantics, a
grep-able shutdown-phase log contract). A 2026-09-16 spike
(`doc/planning/2026-09-16_composition-spike-verdict.md` in go-appkit)
mapped the composition: config mapping is 6/7 identity, the only
missing seam is the listener — with it, appkit delegates `http.Server`
ownership to httputil and drops ~30 lines of parallel plumbing; without
it, composition stays blocked (async bind errors vs the synchronous
contract). Core's deferred TLS option is gated on the same seam
(appkit can only call `StartTLS` after it owns an injectable listener).
