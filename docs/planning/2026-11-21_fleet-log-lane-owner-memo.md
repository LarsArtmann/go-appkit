# Owner memo: the fleet log lane

**Date:** 2026-11-21 · SUPERB v17 leg (c), task T7 · DECISION REQUESTED (the
mechanism SHIPPED today regardless — see "What already landed").

## The question

The fleet's reviews name a logging opinion collision: appkit's built-in logger
is charmbracelet/log **v1** (`github.com/charmbracelet/log v1.0.0`), while
several services and new deps live on the charm.land/log **v2** lane (and some
consumers prefer plain `slog` handlers — dnsblockd names the lane mismatch as
an adoption blocker; go-compliance/testing flag the same friction). Which lane
is the fleet's?

Options:

1. **Stay on charmbracelet/log v1 as the appkit default** (status quo).
2. **Adopt charm.land/log/v2 as the appkit default** (breaking the Logger
   provenance; every existing consumer's log output/appearance may change).
3. **Flex via the handler seam** (RECOMMENDED): appkit's default stays v1, and
   any slog.Handler drops in per service — the handler IS the lane.

## What already landed (T7.1/T7.3–T7.5, 2026-11-21)

`ServiceConfig.LogHandler slog.Handler` — supply any `slog.Handler` and the
service (and its default-stack Logging middleware) emits every line through it;
`LogLevel`/`LogFormat` are ignored in that mode; the lifecycle, middleware
stack, and the grep-able shutdown-line contract are unchanged (pinned:
`loghandler_test.go`). charmbracelet/log's `Logger` already implements
`slog.Handler`, which is exactly why the seam is one field.

Under option 3 the default never flips until the fleet's center of gravity
does — and a consumer on the v2 lane (or on plain slog) is one line away:

```go
appkit.NewService(appkit.ServiceConfig{
    LogHandler: myV2OrSlogHandler,
    // ...
})
```

## The ruling requested

One sentence in the decision log: "the fleet log lane is X". Recommendation:
option 3, with a re-open trigger — when a MAJORITY of new fleet services pick
the v2 lane through handlers in a census, the default follows them (a breaking
Logger-provenance change, then a major-window decision).

## Non-goals

- No zap/logrus bridges shipped in-tree; the seam accepts any `slog.Handler`,
  which is the entire bridge surface the ecosystem needs.
- The charmbracelet/log v1 dependency stays in core either way (it is the
  built-in default), so option 3 costs consumers nothing.
