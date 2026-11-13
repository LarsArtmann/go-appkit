# Archived outgoing asks

Closed history — every file here is a RESOLVED outgoing ask (filed-and-shipped,
posted, mooted, or dropped with a recorded reason). The live filing queue is the
parent directory. Do not re-open these without a concrete provenance need.

## Archive manifest (2026-11-13 docs-health sweep)

| File                                                                  | Classification | Deciding reason                                                                                                                             |
| --------------------------------------------------------------------- | -------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `2026-09-16_upstream-asks-gosse-httputil.md`                          | ARCHIVE        | RESOLVED 2026-10-13/25 without filing (Draft 1 moot — `ReplayFiltered` exists upstream; Draft 2 shipped as httputil `LoggingFromContext` v1.6.0; Draft 3 appkit-side) — resolution banner inline |
| `2026-09-20_upstream-ask-gohealth-recorder-sentinel.md`               | ARCHIVE        | RESOLVED — filed as go-health#7, CLOSED COMPLETED upstream (`ErrRecorderNotApplicable`, go-health v0.8.x), adopted + pinned by appkit health v0.3.0 |
| `2026-09-29_upstream-ask-golicenses-current-toolchain.md`             | ARCHIVE        | RESOLVED-WONT 2026-11-13 — dropped per its own drop rule: premise broke (BuildFlow license-check 13/13 since the 2026-10-07 probe), verdict annotated inline |
| `2026-10-25_upstream-ask-httputil-v160-replace.md`                    | ARCHIVE        | RESOLVED 2026-10-25 — httputil v1.7.0 shipped replace-free, [#43](https://github.com/LarsArtmann/httputil/issues/43) CLOSED COMPLETED (banner inline) |
| `2026-11-03_consumer-heads-up-servertiming-cqrs-htmx.md`              | ARCHIVE        | POSTED 2026-11-06 as cqrs-htmx#43 (dup #44 closed); posting record banner inline                                                            |
| `2026-11-03_consumer-heads-up-servertiming-papdashboard.md`           | ARCHIVE        | POSTED 2026-11-06 as PapDashboard#6 (dup #7 closed); posting record banner inline                                                           |

Left in the parent (still open): samberdo-lazy-healthcheck-docs (FILE-WORTHY,
user-gated), gocqrslite-gracefulclose (DRAFT, user-gated), goflightrecorder-cooldown
(DRAFT, user-gated), the two 2026-11-06 cqrs-lint F032 memos (filing owner-gated).
