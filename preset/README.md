# appkit/preset

One-call composition of the go-appkit family: `preset.New` boots a service
with the always-on spine plus opt-in legs, driven by four frozen environment
switches with explicit-code-wins precedence.

## What joins the service

**Spine (always on):** core `appkit.Service` (lifecycle, middleware stack,
`/health` family), `systemd.Install` sd_notify (no-op outside systemd), health
probe + real-time dashboard (`/healthz` `/readyz` `/startupz`), pretty error
pages (opt-out via env), `appkit.RegisterClassifications()`, and the version
stamp (`GET /version` + build-info label) when `Config.Version` is set.

**Legs (env-driven, code overridable):** see the grammar table below.

## Env grammar (frozen)

| Switch              | Values               | Absent | Meaning                                              |
| ------------------- | -------------------- | ------ | ---------------------------------------------------- |
| `APPKIT_OTEL`       | `on` \| `off`        | off    | otel tracing: Setup + middleware + shutdown flush    |
| `APPKIT_REALTIME`   | `on` \| `off`        | off    | SSE hub mounted at `/events` (configurable pattern)  |
| `APPKIT_ERRORPAGES` | `on` \| `off`        | on     | spine member — opt out with `off`                    |
| `APPKIT_SECURITY`   | `none`\|`standard`\|`strict` | none | security profile (see below)               |

Invalid values fail fast at `New` with a Rejection naming the variable.

**Explicit code wins over env, both directions:** a nil leg field in `Config`
means the env switch decides; a non-nil leg decides itself via its `Enabled`
field and the env var is ignored for that leg.

## Security profiles

- `standard`: `security.SecurityHeaders` (production posture) + `BodyLimit`
  (1 MiB default, configurable).
- `strict`: standard + `security.OriginCheck` — `Config.Security.Origins` is
  REQUIRED and boot rejects without it.

## Quick start

```go
svc, err := preset.New(preset.Config{
    Addr:    ":8080",
    Version: "1.2.3",
})
if err != nil {
    log.Fatal(err)
}
// Register routes on svc.Mux — specific patterns win over the error-pages
// catch-all per stdlib ServeMux precedence.
svc.Mux.HandleFunc("GET /hello", func(w http.ResponseWriter, _ *http.Request) {
    w.Write([]byte("hello"))
})
if err := svc.Run(ctx); err != nil {
    log.Fatal(err)
}
```

Boot with telemetry and the standard security profile:

```bash
APPKIT_OTEL=on APPKIT_SECURITY=standard OTEL_EXPORTER_OTLP_ENDPOINT=https://otel.example:4317 ./myapp
```

The preset never invents credentials: `Config.Metrics` passes through to
`appkit.ServiceConfig.Metrics` verbatim (nil disables the metrics surface).

## Build & verify

```bash
cd preset && GOWORK=off GOTOOLCHAIN=go1.27.2 go test ./... -race -count=1
```

vet/build with the same prefix; lint from this directory
(`golangci-lint run ./...`, one module at a time — concurrent runs race on the
shared build cache).

## Design

Grammar freeze, posture table, conflict-matrix plan:
`docs/planning/2026-11-06_appkit-preset-design-note.md` (repo root).
