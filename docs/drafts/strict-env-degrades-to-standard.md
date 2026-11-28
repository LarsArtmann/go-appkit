# `APPKIT_SECURITY=strict` via env silently boots the standard profile

## Why

The env grammar documents `APPKIT_SECURITY none|standard|strict` (`preset/doc.go:12`), and `resolveLegs` honors it: `parseSecurityProfile` returns strict, `legs.hasSecurity = legs.security != SecurityNone` becomes true (`preset/env.go:102`). But the profile value never reaches the middleware builder, so a strict boot and a standard boot are byte-identical except for the extra boolean.

## Symptom (runtime, 2026-11-24)

Repro module on preset v0.1.1 (go-appkit v0.12.0, security v0.2.0): boot, then `GET /healthz` with `Origin: https://evil.example`.

| mode                                                                             | result                                         |
| -------------------------------------------------------------------------------- | ---------------------------------------------- |
| `APPKIT_SECURITY=strict`, no `Config.Security`                                   | `200 {"status":"pass",...}`                    |
| `Config.Security{Enabled: true, Profile: strict, Origins: [https://ok.example]}` | `403 {"error":{"code":"ORIGIN_REJECTED",...}}` |
| no security leg                                                                  | `200`                                          |

Same intent (strict), two wiring paths, opposite outcomes. The env path arms the leg but drops strict on the floor.

## Root cause (source)

At `5897476` (origin/master), `preset/preset.go:95-97`:

```go
// skip-validate
// preset/preset.go:95 (applySecurity call)
if legs.hasSecurity {
    if err := applySecurity(&svcConfig, cfg.Security); err != nil {
```

It passes `cfg.Security`, which is nil on the env path. `securityMiddlewares` then defaults the profile (`preset/security.go:20-22`):

```go
// preset/security.go:20 (securityMiddlewares)
profile := SecurityStandard
```

so the strict-requires-Origins Rejection (`preset/security.go:38-43`) never fires and `OriginCheck` is never appended (`preset/security.go:56-58`). `legs.security` is computed and then unused for middleware selection.

## Fix

- Pass the resolved profile through: either `applySecurity(&svcConfig, cfg.Security, legs.security)` with `securityMiddlewares` taking the profile as a parameter, or synthesize a `SecurityConfig{Profile: legs.security}` when `cfg.Security == nil`.
- On that path, strict with no Origins should boot-fail with the same Rejection the config path produces today.
- While there: `Config.Security{Enabled: false, Profile: strict}` is also silent (`preset/env.go:105-114` — `hasSecurity = Enabled && profile != none`, the profile is parsed then dropped). Worth an error or a documented ruling.

## To verify

- [ ] `APPKIT_SECURITY=strict` without Origins fails at `preset.New` with `preset.strict_origins_missing`
- [ ] `APPKIT_SECURITY=strict` with `Config.Security.Origins` set rejects a cross-origin request (403 `ORIGIN_REJECTED`)
- [ ] `APPKIT_SECURITY=standard` behavior unchanged (headers + body limit, no OriginCheck)
- [ ] repro table above flips row 1 to 403

Repro module: `/tmp/preset-repro` (2026-11-24; re-create from the table if purged — it is ~60 lines against preset v0.1.1).

💘 Generated with Crush
