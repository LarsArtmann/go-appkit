# appkit/config

Env-first application configuration from struct tags: one `Load` call binds
defaults, an optional YAML/JSON file, the environment, and flag results onto a
plain struct — with classified [go-error-family] failures that name the exact
field path.

It exists because the LarsArtmann fleet kept hand-rolling the same
`envOr`/`envBool`/`envDurationOr` pattern (22 config packages and 25 ad-hoc
`os.Getenv` parsers across 39 Go apps at the 2026-11-06 audit). This module is
that pattern as a library. It is deliberately NOT a config framework: no
providers, no watchers, no interpolation — see [Non-goals](#non-goals) and the
design note (`docs/planning/2026-11-06_02-30_appkit-config-design-note.md` in
the repo) for the koanf adopt-or-decline verdict (declined: fleet reality is
flat env-first structs, not tree-shaped multi-source merges).

## Quick start

```go
type Config struct {
    Addr     string        `default:":8080" env:"APP_ADDR"`
    LogLevel string        `default:"info" env:"APP_LOG_LEVEL"`
    APIKey   string        `env:"APP_API_KEY,required"`
    TTL      time.Duration `default:"60s" env:"APP_TTL"`
    Hosts    []string      `env:"APP_HOSTS"`
}

cfg, err := config.Load[Config](config.WithFile("app.yaml"))
if err != nil {
    log.Fatal(err) // one Rejection listing every field problem
}
```

Run the runnable version: `cd example && GOWORK=off GOTOOLCHAIN=go1.27.2 go run .`

## Precedence

```
flag > env > file > default
```

Fixed and tested as a matrix. Every layer parses from its own raw string
through the SAME parser — a bad default is caught at Load time, never silently
coerced (the koanf `WeaklyTypedInput` trap).

- **default** — the `default:"..."` tag, parsed exactly like an env value.
- **file** — `WithFile(path)`: `.yaml`/`.yml` (go-faster/yaml) or `.json`
  (stdlib). Keys are lowercased field names; nested sections use dotted paths
  (`nested.addr`). Strict: a file key no tagged field declares fails Load —
  typos surface at boot, not in production.
- **env** — the process environment by default; empty values count as unset
  (the fleet's `envOr` convention). `WithLookup` replaces the source for tests.
- **flag** — cmdguard stays the flag engine: after parsing, pass
  `config.WithValues(map[string]string{"Addr": *addrFlag})` (field name,
  case-insensitive, dotted path for nested fields; empty values count as
  unset). Highest precedence.

## Tag grammar

| Tag                                | Meaning                                                               |
| ---------------------------------- | --------------------------------------------------------------------- |
| `env:"NAME"`                       | bind environment variable NAME (layer precedence applies)             |
| `env:"NAME,required"`              | unset/empty after every layer is a Rejection                          |
| `env:"PREFIX"` (on a struct field) | child env keys are `PREFIX_CHILD`; child tags carry only their suffix |
| `default:"..."`                    | value used when env and file are absent                               |

Supported scalar types: `string`, `bool`, all `int`/`uint` widths, `float32/64`,
`time.Duration`, `[]string` (comma-separated, whitespace-trimmed). A tagged
field of any other type fails Load with `unsupported_type` — no silent skips.
Fields WITHOUT an `env` tag are never touched, which is what makes embedding
`appkit.ServiceConfig` safe: only your tagged fields bind.

Embedded (anonymous) struct fields are walked transparently — their leaves
bind as promoted top-level fields.

## Errors

`Load` returns a nil config and a single go-error-family **Rejection**
(`config.invalid`) wrapping one `*config.FieldError` per problem:

```go
_, err := config.Load[Config]()
var fe *config.FieldError
if errors.As(err, &fe) && fe.Kind == config.KindMissingRequired { ... }

for _, fe := range config.FieldErrorsOf(err) {
    // fe.Field "APIKey", fe.Key "APP_API_KEY", fe.Kind KindMissingRequired
}
```

Kinds: `missing_required`, `parse_failed`, `unsupported_type`, `file_read`,
`file_decode`, `unknown_extension`, `unknown_field`, `validation`.

## Discoverability

Push-model tags make the key set a fact of the type:

```go
keys, _ := config.Keys[Config]() // ["APP_ADDR", "APP_API_KEY", ...] sorted
```

Table tests iterate `Keys` to prove every key is wired (set junk, expect a
`parse_failed`) — the push-model replacement for koanf's
`AssertEnvOverridesReachable` debugging helper.

## Non-goals (v0.1.0)

Hot-reload/watch (battery F4, demand-gated), interpolation, remote sources,
multi-file merge, dotenv loading, tree-shaped provider registries. Compose
`os.ReadFile` + `Load` in main if you need dotenv; the module stays flat.

## Build & verify

```bash
cd config && GOWORK=off GOTOOLCHAIN=go1.27.2 go test ./... -race -count=1
# vet/build with the same prefix; lint: cd config && golangci-lint run ./...
```

[go-error-family]: https://github.com/LarsArtmann/go-error-family
