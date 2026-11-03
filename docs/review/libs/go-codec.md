# go-codec — Utilization Deep Dive

> **Point-in-time review · 2026-11-03 · go-appkit @ master (post-SUPERB v10)**
> Series: part of the 12-library fleet audit (`docs/review/libs/index.html`). First dedicated
> report. The content-negotiation pattern reference is
> `docs/planning/2026-10-01_go-daemon-and-go-aichat-integration.md` (adopt-the-pattern row).

> **⚠ OWNER RULING 2026-11-03 (post-audit errata — supersedes the "should NOT use now" framing below).**
> The owner clarified the repo's mission: go-appkit exists to make his apps smartly compose the
> family's libraries. Under that mission, go-codec adoption is IN-SCOPE — the question is WHERE
> the seam goes, and the honest answer stands: appkit has no payload-encoding boundary today
> (errorpages' JSON body and the docs catalog don't benefit from a codec), so the mission-correct
> adoption is to BUILD the boundary on go-codec: the W3-httpx content-negotiation battery (B5)
> (`codec.ForEncoding` + Accept q-value ranking, go-daemon's `negotiation.go` as reference), plus
> direct requires wherever appkit gains its first encode surfaces. Zero-direct-use is a true
> statement about today's seams, not a verdict against the library. The recorded plan already
> routed exactly this; the ruling promotes it from "someday, demand-gated" to planned work.

## Verdict

**NOT USED — deliberately. Zero imports across all 12 modules; the library rides as an
indirect (cqrs v0.3.3 = latest tag; docs/integration carry v0.3.2 floors) because
go-cqrs-lite/command/v4 builds on it. Appkit never encodes event payloads itself — the cqrs
wrapper delegates payload codecs to go-cqrs-lite, which is exactly the layering the family
agreed on. The recorded future surface: battery B5 (content negotiation in the W3 httpx
module) is pre-agreed to be built ON go-codec, not on go-daemon's copy of the pattern.**

| Fact              | Value                                                                                                | Evidence                                                    |
| ----------------- | ---------------------------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| Direct imports    | 0 of 12 modules                                                                                      | `rg 'go-codec' --glob '*.go'` → empty                       |
| In-graph versions | cqrs v0.3.3 (= latest); docs v0.3.2 + integration v0.3.2 (MVS union resolves v0.3.3 for repo builds) | `cqrs/go.mod:53`, `docs/go.mod:16`, `integration/go.mod:87` |
| Chain             | cqrs → go-cqrs-lite/command/v4 → go-codec                                                            | `go mod why -m` (verified 2026-11-03)                       |
| Upstream state    | v0.3.3 (2026-10-29, decode-only additive); HEAD 1 commit ahead                                       | `~/projects/go-codec/CHANGELOG.md`                          |

## 1. Capability inventory

An event-payload codec SDK (`~/projects/go-codec/`, single package):

- **Codecs**: canonical RFC 7049 `CBORCodec` (recommended), `CBORCompactCodec` (RFC 8949
  deterministic, INCOMPATIBLE with CBORCodec output — new stores only), `JSONCodec`
  (+`encoding/json/v2` build-tag path), `RawCodec`.
- **Envelope**: `WrapEncode`/`UnwrapDecode` (self-describing "gcdc" magic) +
  `DecodeEnvelopeOrLegacy[T]` for migration windows.
- **Auto-detect**: `AutoDetect`/`AutoDetectDebug` with a stable `Reason` contract.
- **Transcode/streaming/pooling**: `TranscodeToJSON` (schema-free CBOR→JSON), NDJSON +
  CBOR stream encoders, `EncodePooled` (≤1 MiB buffer pool), `Size` probing.
- **COSE (RFC 9052/9053)**: Sign1/Encrypt0 marshal/unmarshal + generic protected-header
  builder — the primitives the cqrs-lite signing/encryption modules consume.
- **Observability**: `ObserveCodec` wrapper with metrics hooks; error-family typed
  sentinels (never match strings).
- Gotchas: `toarray`/`keyasint` struct tags ARE wire format (append-only evolution);
  `time.Time` payloads must be UTC; `EncodePooled` bytes valid only in-callback.

## 2. Capability vs applicability

| Capability                      | Status                                | Reason                                                                                                                                                                                                                                                       |
| ------------------------------- | ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Payload codecs in appkit/cqrs   | ⚪ N/A — layering by design           | cqrs delegates payloads to go-cqrs-lite (which requires go-codec); appkit's wrapper surface stays command/query + projections + DLQ                                                                                                                          |
| Content negotiation (future B5) | ⚪ N/A today — adoption path recorded | `docs/planning/2026-10-01_go-daemon-and-go-aichat-integration.md` pre-agrees: build negotiation on go-codec in the W3 httpx module, using go-daemon's `ResolveRequestCodec`/`ResolveResponseCodec` as the reference implementation — do not import go-daemon |
| COSE / envelope / autodetect    | ⚪ N/A                                | Signing/encryption are cqrs-lite module concerns (deferred EventConfig opt-ins, TODO_LIST:52)                                                                                                                                                                |
| Version currency                | 🟢                                    | v0.3.3 in cqrs = newest tag; v0.3.2 floors in docs/integration self-heal at the next re-pin                                                                                                                                                                  |

## 3. Why we SHOULD NOT use it now (and when that flips)

1. **No encoding boundary exists in appkit.** The framework moves events through
   go-cqrs-lite; codecs behind an interface it does not own. Adding go-codec to any appkit
   module today would create a second codec decision next to cqrs-lite's.
2. **The indirect pin is structural**: go-cqrs-lite/command requires it, and MVS keeps it at
   the newest tag — graph hygiene is already optimal.
3. **When it flips:** battery B5 (httpx content negotiation) — at design time the answer is
   `codec.ForEncoding` + `ResolveRequestCodec`-style q-value ranking on top of go-codec,
   with the CBOR-compatibility gotchas (compact vs canonical) called out in the module docs.
   The signing/encryption EventConfig opt-ins, if ever demanded, ride go-codec COSE through
   cqrs-lite — still no direct appkit import.

## 4. Recommendations

1. Keep the posture; B5's "on go-codec" decision is already recorded in the go-daemon
   integration doc — link it from the W3 battery row when that row is next touched
   (annotation-level).
2. Version hygiene: docs/integration floors move at the next integration/docs re-pin train —
   no action needed.

## 5. How to verify

```bash
cd /home/lars/projects/go-appkit
rg -n 'larsartmann/go-codec' --glob '*.go'                # empty → zero imports
grep 'go-codec' cqrs/go.mod docs/go.mod integration/go.mod
cd cqrs && GOWORK=off GOTOOLCHAIN=go1.27.2 go mod why -m github.com/larsartmann/go-codec
rg -n 'go-codec' docs/planning/2026-10-01_go-daemon-and-go-aichat-integration.md | head -3
git -C /home/lars/projects/go-codec tag --sort=-v:refname | head -1   # v0.3.3
```
