# Outgoing upstream feedback — drafts ready to file (FILING IS USER-GATED)

**Created:** 2026-09-16 · **Status:** drafted, verified against pinned sources, NOT filed.
**Rule:** verify-before-filing + github-voice apply; each draft cites the verified source it is built on.

---

## Draft 1 — go-sse: dedup-aware `ReplayFiltered` (F121)

**Repo:** github.com/larsartmann/go-sse (v0.6.0) · **Where:** feature request

**Problem:** appkit's `realtime.Handler` re-implements replay filtering for
live-dedup. The handler must (a) subscribe to the live broadcaster BEFORE
reading the store (closing the snapshot-to-subscription gap) and (b) skip
live events the replay already delivered. Upstream `ReplayFiltered` has no
dedup handle, so the consumer collects replayed IDs itself and filters in the
live loop.

**Ask:** a dedup-aware replay variant, e.g. `ReplayFilteredDedup(store,
lastID, filter, seen func(id string) bool)` or a returned
`ReplayedIDs map[string]struct{}` helper, so the subscribe-before-replay +
dedup pattern becomes one upstream call.

**Reference implementation:** go-appkit `realtime/handler.go`
(`replayMissedEvents` + `forwardLive` dedup block), production-scarred in the
cqrs-htmx journal-replay E2E (`integration/integration_test.go`:
`TestJournalBackedReplayThroughAppkitService` — exact missed-suffix replay +
dedup under live broadcast interleave).

**Acceptance:** reconnect with concurrent broadcasts delivers each event
exactly once; filtered-out events don't count as delivered.

---

## Draft 2 — httputil: `Logging` completion line with request context (F122)

**Repo:** github.com/larsartmann/httputil (v1.2.0) · **Where:** feature request

**Problem:** the `Logging` middleware emits the request-completion line
WITHOUT the request's context, so slog handlers that decorate records with
trace/span IDs (appkit/otel's `TraceHandler`) cannot correlate the one log
line every ops runbook greps for. Only handler-level logs correlate today.

**Ask:** a context-aware emit variant (`LoggingConfig.WithRequestContext`) or
the emit call switched to `slog.InfoContext(r.Context(), ...)` where the
request already carries the request-ID/tracing context — the pattern-propagation
work (v1.2.0) already demonstrates the same fork-aware discipline.

**Reference:** go-appkit `otel/logging.go` `TraceHandler` (the decorator that
stays uncorrelated for exactly this line), documented as a known limitation in
`otel/README.md` and AGENTS.

**Acceptance:** a request passing through `Logging` + `TraceHandler` shows
`trace_id` on the completion line in an in-memory-handler test.

---

## Draft 3 — F2 timing battery sketch (stays appkit-side until F2 graduates)

Request-ID-seeded handler logger + `X-Response-Time` header (CV
`internal/middleware/timing.go`, 116 LOC, verified port-ready):

```go
// middleware: start := time.Now(); label request start into the logging
// context; defer sets X-Response-Time once the handler returns.
func Timing(logger *slog.Logger) func(http.Handler) http.Handler
```

**Sequencing:** lands with (or after) Draft 2 — the completion line and the
header should carry the SAME duration source, or dashboards will disagree.

---

## Cross-references (2026-09-20 health-stack train)

- Draft 2 (Logging completion-line correlation) re-confirmed during the
  health-stack composition work: the integration E2E
  (`integration/health_stack_test.go`) asserts readiness surfaces but cannot
  assert trace-correlated completion lines for the same reason documented in
  `otel/doc.go` — the limitation is stable, not transient.
- New sibling drafts from the same session:
  `2026-09-20_upstream-ask-gohealth-recorder-sentinel.md` (recorder sentinel
  on the function-path constructors, includes a secondary `Probe.Evaluate`
  cache-publication godoc ask) and
  `2026-09-20_upstream-ask-samberdo-lazy-healthcheck-docs.md` (unbuilt lazy
  services report healthy). Filing remains USER-gated (Gate Q3).
