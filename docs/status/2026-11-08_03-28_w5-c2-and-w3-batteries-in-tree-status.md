# Status: W5-C2 contract + W3 batteries — in-tree, pre-release-train

**Date:** 2026-11-08 03:28 CET · **Session:** whole-TODO-list execution mandate, interrupted mid-train for status · **State:** code complete + suites green + lint 0 in-tree (daemon auto-committed at `582a3e1`); NO tags created, NO master push, release train NOT started.

## a) What the user asked for and got

Execute the entire actionable TODO list. Delivered so far (all verified in-tree):

1. **20.2a cqrs folded-contract half** — `cqrs/folded.go` (+tests, 6): `FoldedEvent` / `FoldedNotifier` / `NotifyFolded` (non-blocking send, `DroppedNotifications` counter, nil-channel disable, failed-fold-no-notify, fold-order preservation). Suite `-race` green, golangci 0, README section landed.
2. **20.2b realtime folded-contract half** — `realtime/folded.go` (+tests, 7): `FoldedSignal` interface (cqrs.FoldedEvent satisfies structurally — zero cross-imports), generic `FoldedSource` pump (snapshot-strictly-after-signal, full-state broadcasts, ctx/channel-close lifecycle, marshal-failure self-heal). Suite green, lint 0, README + doc.go contract sections + C1-dedupe decision recorded.
3. **Test-helper migration** — flightrecorder local `httpGetURL` → `testkit.HTTPGetURL`; integration `freeAddr`/`e2eGet` deleted → `testkit.FreeAddr`/`HTTPGetURL` (~35 sites); integration suite hermetic `-race` green; drain-walker + sse-protocol guards green. **Deliberate deviation:** realtime keeps its 15-line local helper (protects its 4-dep standalone graph).
4. **httpx B3** — `WrapWriter` + new `httpx/httpxtest` subpackage: `AssertForwardsInterfaces`/`ForwardingFindings`; the buggy plain-embedding/no-op-liar/controller-opaque shapes are proven to FAIL (the discrimination acceptance).
5. **httpx B4** — `Conditional` (go-etag/server adoption, strong sha256 default, `ETagFunc`/`SkipConditional` options) + httpx-side If-Modified-Since layer with RFC precedence; CV matrix ported and green (strong/weak/comma/*, IMS, precedence, HEAD, skip).
6. **httpx B5** — `Negotiate` (RFC 9110: q/specificity/order ranking, q=0 refusal, ""-on-unsatisfiable, most-specific-range-governs) + the go-codec seam (`RequestCodec` 415-on-unsupported, `ResponseCodec` refusal-substitution semantics). Tests found and fixed 3 real ranking bugs. go-codec v0.3.3 direct require.
7. **httpx B7** — `WriteDeadline(extension, prefixes...)`: real-server deadline-extension E2E + control + loud-500 ordering regression (503→500 override recorded).
8. **httpx hygiene** — fleet `_test.go` exclusion union added to its .golangci.yml (v0.1.0 onboarding gap), ireturn allow for `go-codec.Codec`, lint **0 issues**, README/doc.go battery docs.

## b) Todo list state

Marked in the session tracker: recon ✓, 20.2a ✓, 20.2b ✓, helper-migration ✓, B3 ✓, B4 ✓, B5 ✓, B7 ✓. Remaining queue (mine): CHANGELOGs (was mid-task at interrupt), integration discriminator E2E + re-pin (blocked on tags), systemd small-backlog train, exit-criteria refresh, the three-tag release wave, truth-sync + fleet verification. Owner-gated/watchlist rows untouched by design.

## c) Biggest findings/decisions

- **Negotiate semantics bugs caught by tests:** absent-Accept vs wildcard must differ in eligibility; the MOST SPECIFIC matching range (not the highest q) governs one offer's standing; error-family's `WithHTTPStatus` is needed because `Wrapf` shadows family-415 with family-400.
- **realtime standalone ruling:** no core import for a test helper (dep-graph cost >> 15-line dedup).
- **go-etag layering:** `Conditional` adopts upstream's buffered ETag writer and owns only the IMS/precedence layer — no wheel rebuilt.
- **httpx lint-config gap:** v0.1.0 shipped without the fleet test-exclusion block; now consistent.

## d) Biggest risk right now

The release train is the unexecuted half: tags must be cut only after CHANGELOG datings + guards (daemon moves tags; tag LAST), dependabot-parity may demand go-etag/go-codec entries, cqrs-lint + API-break diffs (additions-only → minor) must run per module, and the integration re-pin + discriminator E2E land on master only after proxy indexing (release-ordering gotcha #6).

## e) Next step

Awaiting owner answers (§g). On "go": CHANGELOGs → pre-tag-checks (`cqrs/v0.9.0 realtime/v0.2.0 httpx/v0.2.0`) → push tags → proxy wait → gosum-canonical integration → master push with integration re-pin + discriminator + documentedPins bump → systemd train → exit-criteria → truth-sync.

## f) Anything unclear

Nothing blocking that research can resolve; the three owner calls below are genuine decision gates.

## g) Questions for the user

1. Authorize the three-tag wave (cqrs v0.9.0, realtime v0.2.0, httpx v0.2.0) — tags + push this session?
2. Realtime-standalone deviation (no testkit import) — confirm or overrule?
3. systemd v0.2.0 train (Install returns Hooks + watchdog README note) — same wave or own train?
