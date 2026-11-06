# go-etag — Utilization Deep Dive

> **Point-in-time review · 2026-11-03 · go-appkit @ master (post-SUPERB v10)**
> Series: part of the 12-library fleet audit (`docs/review/libs/index.html`). Supersedes the
> go-etag section of `docs/research/2026-10-11_sibling-library-utilization-audit.html`
> (newer and narrower — the library has GROWN since that audit: client + metrics modules).

## Verdict

**NOT USED — deliberately, and the posture remains exemplary. Zero imports across all 12
modules; the library is an intentional indirect (entitytag + server v0.6.1 in 8 go.mods)
because httputil registers go-etag's error codes in its translation table — no appkit code
path generates or evaluates an ETag. The default stack has no conditional-GET middleware.**
What changed since the 2026-10-11 audit: go-etag v0.6.0 split into FIVE modules and grew a
client-side conditional-GET `RoundTripper` plus a metrics package — which widens the
adoption surface beyond the server-side battery B4 story on file.

| Fact             | Value                                                                         | Evidence                             |
| ---------------- | ----------------------------------------------------------------------------- | ------------------------------------ |
| Direct imports   | 0 of 12 modules                                                               | `rg 'go-etag' --glob '*.go'` → empty |
| In-graph version | entitytag v0.6.1 + server v0.6.1 (= latest), indirect × 8                     | `go.mod:22-23` + satellites          |
| Root module      | deprecated shim (removal at v1.0.0)                                           | `~/projects/go-etag/deprecated.go`   |
| Graph hygiene    | integration's exclude block fences pre-extraction root versions v0.1.1-v0.5.0 | `integration/go.mod:174-177`         |
| Modules          | 5: root shim, entitytag (zero-dep), server, client, metrics                   | `~/projects/go-etag/*/go.mod`        |

## 1. Capability inventory (what the library now offers)

- **entitytag** (zero deps): RFC 7232 `ETag` type, strong/weak comparison, wire parsing
  (`ParseETag`/`ParseETagList`), `MatchesIfNoneMatch`/`MatchesIfMatch`.
- **server**: `etag.New(ETagConfig)` middleware — buffers GET/HEAD bodies up to
  `MaxBufferSize`, emits the tag, answers `If-None-Match` with 304; `SkipIfPresent`
  (adopt handler tags), `Skip` predicate, observability hooks
  (`OnError`/`OnETagGenerated`/`On304`/`OnBufferOverflow`); typed error codes over
  go-error-family (`Code`/`Domain` API + `RegisterErrorClassifications`).
- **client** (NEW since the last audit): `etagclient.NewTransport` — conditional-GET
  `http.RoundTripper` with replay-cache (`Stats`, `MaxEntries`, `MaxBodyBytes`),
  RFC 9111 §4.3.5 304-freshening, unsafe-method invalidation, `KeyFunc` for
  credential-sensitive cache keys.
- **metrics** (NEW): `Attach(cfg)` installs atomic counters (`Generated`/`NotModified`/
  `BufferOverflows`) + `HitRatio()` without touching user hooks.

## 2. Capability vs applicability

| Capability                          | Status                                | Reason                                                                                                                                      |
| ----------------------------------- | ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| server middleware (conditional GET) | ⚪ N/A today — adoption path recorded | Battery B4 (W3 httpx) explicitly "promotes the unused go-etag indirect dep" (TODO_LIST:50); no consumer asks for HTTP caching semantics yet |
| client `RoundTripper`               | ⚪ N/A today — NEW adoption surface   | Relevant to the planned W4 `polite` outbound client battery, which has no in-framework consumer today                                       |
| metrics `Attach`                    | ⚪ N/A today                          | Only meaningful once the server middleware is wired                                                                                         |
| entitytag primitives                | ⚪ N/A                                | No appkit code manipulates validators                                                                                                       |
| Version currency                    | 🟢                                    | v0.6.1 everywhere = newest tag; excludes block keeps MVS clean                                                                              |

## 3. Why we SHOULD NOT use it now (and when that flips)

1. **No conditional-GET story exists or is demanded.** The errorpages conditional-GET idea
   was closed "won't implement — on consumer demand only" (08-53 audit); battery B4 owns the
   promotion path. Adding ETag middleware to the default stack would change wire behavior
   (304s, buffered bodies) for every consumer without a demand signal.
2. **The indirect pin is load-bearing, not waste.** httputil's error-translation table
   registers go-etag codes; the indirect is what keeps classification templates complete.
   Removing it would degrade error templates, not clean the graph.
3. **When it flips:** (a) a consumer asks for HTTP caching semantics → battery B4 wires the
   server module inside the W3 httpx home (not the default stack); (b) the `polite` outbound
   client battery (W4) is designed → evaluate `etagclient.NewTransport` as the caching layer
   with a credential-aware `KeyFunc` (the README's credential warning is the first review
   item); (c) either way, `metrics.Attach` comes along for observability.

## 4. Recommendations

1. Keep the posture and the B4 routing — re-verified current (TODO_LIST:50).
2. UPDATE the adoption note to mention the enlarged surface: battery B4's text says
   "promotes the unused go-etag indirect dep" — it should name the five-module split and
   add the client-module option to W4's `polite` sketch. (Annotation-level; no code.)
3. At the next go-etag release, confirm the root shim removal plan (v1.0.0) does not affect
   the integration exclude block (it fences v0.1.1-v0.5.0; v0.6.x split modules are
   unaffected).

## 5. How to verify

```bash
cd /home/lars/projects/go-appkit
rg -n 'larsartmann/go-etag' --glob '*.go'                    # empty → zero imports
grep -h 'go-etag' go.mod */go.mod | grep -v exclude | sort -u  # entitytag+server v0.6.1 indirect
sed -n '174,177p' integration/go.mod                          # pre-extraction excludes
ls /home/lars/projects/go-etag/                               # entitytag/ server/ client/ metrics/
git -C /home/lars/projects/go-etag tag --sort=-v:refname | head -1  # v0.7.0
```

## Errata (2026-11-06, re-run of the verify block)

- Block verified as written; the only drift is version: go-etag's latest tag is
  now **v0.7.0** (was v0.6.1 at audit time). The grep also surfaces
  integration's exclude-block entries (v0.1.1/v0.3.1/v0.4.0/v0.5.0 — the
  entry lines themselves carry no "exclude" text; the block header does), and
  the sed range still lands on the excludes at lines 173-178. appkit's
  indirect pins remain entitytag+server v0.6.1 (member go.mod changes ride
  release trains).
