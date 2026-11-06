# appkit/preset — Design Note (SUPERB v12, P1)

**Created:** 2026-11-06 · **Status:** GRAMMAR FROZEN (T1.1) · **Plan:**
`docs/planning/2026-11-06_01-00-SUPERB-v12-app-composition-preset-plan.html`
(T1.1–T1.6) · **Feedstock:** session-2 signature research, re-verified via
`go doc` 2026-11-06 against the published tags the module will require.

One-call composition: `preset.New` boots a service with the always-on spine
plus opt-in legs, driven by env switches with explicit-code-wins precedence.
The preset composes SHIPPED batteries only — no new capability, no fat core.

## 1. Spine pieces and their config surfaces

| Piece | Battery call (verified signature) | Always on? |
| ----- | --------------------------------- | ---------- |
| Core service | `appkit.NewService(appkit.ServiceConfig) (*appkit.Service, error)` | YES |
| sd_notify | `systemd.Install(&cfg)` — appends Start/Drain/Shutdown hooks; no-op outside systemd | YES |
| Health probe + dashboard | `health.NewProbe(map[string]health.CheckFunc, ...health.Option) *health.Probe` + `health.Mount(svc.Mux, probe, health.WithDashboard()) (*health.Mounted, error)` → `/healthz /readyz /startupz` + dashboard (no collision with core's `/health` family) | YES |
| Error pages | `errorpages.Mount(svc.Mux, cfg)` — catch-all `/` pretty-404/JSON | YES (spine member; env can opt OUT) |
| Error classification | `appkit.RegisterClassifications()` (idempotent; deliberately not auto-called by core — the preset owns the call) | YES |
| Version stamp | `ServiceConfig.Version` → `GET /version` + build_info label | YES when consumer sets it |
| Metrics | `ServiceConfig.Metrics *MetricsConfig` — **passthrough only**: preset never invents credentials; nil stays nil | passthrough |

Wiring order (chicken-egg): hooks that reference `*health.Mounted` (and the
otel provider, realtime hub) are appended to the config BEFORE
`appkit.NewService`, as nil-guard closures; the pointed-to values are assigned
after their `Mount`/`Setup` calls on `svc.Mux`. Hooks only run at Run/Start or
Shutdown — long after assignment — so the closures are safe. This is the
pattern the core starthooks_test already pins (hooks observe the bound
listener).

Health lifecycle mapping: `Mounted.Start(ctx)` → StartHooks (before serving);
`Mounted.Drain()` → DrainHooks (readiness flips in lockstep with the drain
window); `Mounted.Shutdown(ctx)` → ShutdownHooks. `systemd.Install` appends its
hooks BEFORE preset appends these, so sd_notify READY=1 lands after
`Mounted.Start`? No — hook order is append order; Install is called FIRST
(READY announces the listener, which is correct), then health legs append. The
drain mirror holds: STOPPING=1 (systemd, first) then `Mounted.Drain` —
readiness endpoints report down for the whole drain window either way because
the ready flip happens before DrainHooks run at all.

## 2. Env-switch grammar — FROZEN

| Switch | Type | Values | Absent | Invalid |
| ------ | ---- | ------ | ------ | ------- |
| `APPKIT_OTEL` | tri-bool | `on`, `off` | **off** (zero otel activity) | Rejection at boot |
| `APPKIT_REALTIME` | tri-bool | `on`, `off` | **off** | Rejection at boot |
| `APPKIT_ERRORPAGES` | tri-bool | `on`, `off` | **on** (spine member) | Rejection at boot |
| `APPKIT_SECURITY` | enum | `none`, `standard`, `strict` | **none** (nothing silently joins a stack) | Rejection at boot |

Rules:

1. **Explicit code wins over env, both directions.** Every leg is a pointer in
   `Config` (`*OTELConfig`, `*SecurityConfig`, `*RealtimeConfig`,
   `*ErrorPagesConfig`): nil → the env switch decides; non-nil → the leg's own
   `Enabled bool` (first field) decides and the env var is IGNORED for that
   leg. This mirrors otel's `OTEL_*_EXPORTER=none` posture (explicit code
   wins) already shipped in otel v0.3.0.
2. No per-app custom switches in v0.x — four switches, frozen; more legs join
   only through a minor-version design note.
3. Values are case-sensitive lowercase, no trimming.

## 3. preset.Config sketch (design-first; README copies this)

```go
package preset // import "github.com/larsartmann/go-appkit/preset"

type Config struct {
    // Core passthrough — every field appkit.ServiceConfig has that a
    // consumer plausibly sets. Preset fills defaults, never invents:
    // Metrics, ReadyCheck, Middlewares*, Timeout knobs pass through verbatim.
    Addr        string
    LogLevel    appkit.LogLevel
    LogFormat   appkit.LogFormat
    Version     string
    Metrics     *appkit.MetricsConfig // passthrough; preset never sets creds
    ReadyCheck  func() bool
    Timeouts    TimeoutsConfig // Read/ReadHeader/Write/Idle/Shutdown/DrainDelay

    // Legs — nil means the env switch decides (see grammar table).
    OTEL        *OTELConfig        // nil + APPKIT_OTEL=on → Setup with defaults
    Security    *SecurityConfig    // nil + APPKIT_SECURITY=standard|strict → profile
    Realtime    *RealtimeConfig    // nil + APPKIT_REALTIME=on → hub at /events
    ErrorPages  *ErrorPagesConfig  // nil + APPKIT_ERRORPAGES (absent=on) → Mount

    // Spine knobs.
    HealthChecks map[string]health.CheckFunc // empty map default: probe with no checks
    NoDashboard  bool                        // omit health.WithDashboard
}

func New(cfg Config) (*appkit.Service, error) // builds + wires; consumer then registers routes on svc.Mux and calls svc.Run(ctx)
```

Leg sub-configs: `OTELConfig{Enabled bool; ServiceName, ServiceVersion,
InstanceID string; Options []otel.SetupOption}`,
`SecurityConfig{Enabled bool; Profile SecurityProfile; Origins []string;
BodyLimitBytes int64}`,
`RealtimeConfig{Enabled bool; Pattern string; Options []realtime.Option}`,
`ErrorPagesConfig{Enabled bool; Config errorpages.Config}`.

`New` flow: parse env (invalid → `errorfamily` Rejection, fail-fast before any
listener binds) → resolve each leg (code vs env) → assemble ServiceConfig
(passthrough + hooks) → `systemd.Install(&sc)` → `appkit.NewService(sc)` →
mount health (dashboard) → mount errorpages if on → otel `Setup` + Middleware
if on → realtime hub + Mount if on → assign nil-guard closure targets → return
service. Route registration order is consumer-last: errorpages.Mount's
catch-all `/` at New means later consumer patterns (specific) win per stdlib
ServeMux precedence — documented in the README.

## 4. Posture table

| Environment | OTEL | Security | Realtime | ErrorPages | Notes |
| ----------- | ---- | -------- | -------- | ---------- | ----- |
| dev (all absent) | off | none | off | on | zero-config boot; dashboard + pretty 404s |
| prod | `APPKIT_OTEL=on` | `APPKIT_SECURITY=standard` | per app | on | one env line each; creds via OTEL_* upstream vars |
| strict (browser-facing) | `APPKIT_OTEL=on` | `APPKIT_SECURITY=strict` | on | on | strict adds OriginCheck — Origins REQUIRED in SecurityConfig, else Rejection |
| systemd unit | any | any | any | any | Install is always called; no-op outside systemd |

## 5. Conflict-matrix test plan (P7 suite)

For each of the four switches, all combinations:

| Code leg | Env | Expectation |
| -------- | --- | ----------- |
| nil | absent | default (grammar table) |
| nil | `on`/`standard` | leg active |
| nil | `off`/`none` | leg inactive |
| `Enabled: true` | `off`/absent | ACTIVE — code wins |
| `Enabled: false` | `on` | INACTIVE — code wins |
| nil | `banana` | boot Rejection (fail-fast, message names the var + valid values) |

Leg observables: OTEL → middleware in chain (traceparent response header with
a test sampler / provider shutdown hook ran); Security → response headers
(`X-Content-Type-Options`, HSTS under production) + body-limit 413; Realtime →
`GET /events` streams (sync on `hub.SubscriberCount()` before broadcast — the
SSE house rule); ErrorPages → unknown path serves pretty 404 (HTML/JSON by
Accept).

Spine tests (P2/P3): `/healthz`,`/readyz` 200 + dashboard route 200; graceful
stop runs hooks in order; sd_notify under a `Type=notify` harness is exercised
by the systemd module's own tests (preset relies on Install append-only
contract); full-spine-boots-with-zero-app-code (empty Config, no env).

## 6. Non-goals (inherited from the plan)

No fat core; no app rewrites from this repo; no DI container (samber/do stays
a W4 battery decision); security never silently joins (absent = none); no auth
leg in v0.x (T5.4 owner gate); preset does not wrap `svc.Run` — consumer owns
the blocking call and route registration.

## 7. Module facts (P1 skeleton)

Requires (all PUBLISHED tags): appkit v0.9.0, systemd v0.1.0, health v0.3.0,
errorpages v0.1.2, otel v0.3.0, realtime v0.1.3, security v0.2.0. NO config
require — the preset reads its four frozen env vars directly via os.Getenv
(fail-fast Rejection), and a require nothing imports would not survive
`go mod tidy`. Until this repo's local tags are pushed, local tidy uses the
git-config insteadOf VCS trick (Release Ritual step 4). Floor 1.27.2, LICENSE,
satellite `.golangci.yml` (tagalign tweak only if optional tags appear),
charter `--fix`, dependabot entry, ci.yml matrix slot (test + proxy-smoke at
first tag), README Build & verify, CHANGELOG.
