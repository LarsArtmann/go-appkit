# `NoDashboard: true` panics at boot once go-health >= v0.9 is in the graph

## Symptom

`preset.New(Config{NoDashboard: true, Version: "vtest", ...})` panics:

```text
panic: health: Routes.Version mounts the build-version endpoint, but the probe has no build stamp —
call WithVersion(...) at construction (ldflags/VCS recipe in the README), or leave Routes.Version empty
to keep /version unmounted

github.com/larsartmann/go-health.(*Probe).RegisterRoutes(...)
    github.com/larsartmann/go-health@v0.10.0/handlers.go:160
github.com/larsartmann/go-appkit/health.(*Mounted).RegisterRoutes(...)
```

Reproduced 2026-11-24 against preset v0.1.1 with go-health resolved to v0.10.0 (MVS pulls it via go-health-dashboard v0.13.3). Tuple sensitivity: with go-health pinned at v0.8.1 the same binary boots fine — v0.8.1 has no Version route at all. So every preset consumer is one dashboard bump away from this panic, and the template repo hit it exactly that way.

## Root cause (source)

The preset never stamps the probe. At `5897476` (origin/master), `preset/preset.go:109-111`:

```go
// preset/preset.go:109
probe := health.NewProbe(cfg.HealthChecks)

mounted, err = health.Mount(svc.Mux, probe, dashboardOpts(cfg.NoDashboard)...)
```

With `NoDashboard: true`, `dashboardOpts` returns nil (`preset/preset.go:130-131`), so `appkit/health` `Mount` defaults `probeRoutes: health.DefaultRoutes()` (`health/mount.go:70`), whose Routes include Version. The nil-dashboard branch then registers the probe routes directly (`health/mount.go:115-122`), and go-health >= v0.9 panics at registration when `Routes.Version != ""` on an unstamped probe. The default dashboard path dodges it only because `RegisterRoutes` early-returns into the dashboard (`health/mount.go:116-118`).

The fix slot already exists: `appkit/health.NewProbe` forwards options (`health/probe.go:63-69`), so one line stamps it:

```go
probe := health.NewProbe(cfg.HealthChecks, health.WithVersion(cfg.Version))
```

`cfg.Version` is already a required-ish field (`preset/config.go:38-39` — it stamps GET /version), so the stamp cannot be empty in practice.

## Fix

- [ ] `probe := health.NewProbe(cfg.HealthChecks, health.WithVersion(cfg.Version))` in `preset.New`
- [ ] a boot test with `NoDashboard: true` on a go-health >= v0.9 graph (today nothing in the preset test matrix covers the nil-dashboard registration branch)

## To verify

- [ ] repro flips from panic to boot; `/version` serves the stamped value on the NoDashboard path
- [ ] default dashboard path unchanged

💘 Generated with Crush
