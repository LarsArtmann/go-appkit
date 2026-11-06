# appkit/config — Design Note (F1, C1)

**Created:** 2026-11-06 · **Status:** DECIDED — implements battery F1 as elevated by the
cross-app audit (22 hand-rolled config packages, 25 `os.Getenv` parsers across 39 Go apps)
· **Module:** `github.com/larsartmann/go-appkit/config` (new satellite, package `config`)

## 1. What this is in the domain

The fleet's config layer is a flat struct per app resolved from the environment with
defaults, occasionally from a file, validated by hand. The audit found the SAME ~60-line
pattern (envOr/envBool/envDurationOr + a required-check + a validate func) in 22 repos.
`appkit/config` is that pattern as a library: **declarative struct tags, one Load call,
classified errors.** It is not a config framework — no providers, no watchers (F4
hot-reload stays demand-gated), no pull-model registry.

## 2. The koanf adopt-or-decline verdict — DECLINE

- koanf is the right tool for tree-shaped multi-source merges; exactly ONE fleet repo
  (dnsblockd) has that shape, and it already carries koanf — it does not need appkit to.
- The other 21 config packages and all 25 ad-hoc `os.Getenv` parsers are flat
  env-first structs; a pull-model provider registry would make every consumer MORE
  code, not less (the F1 spec's own discoverability helper exists only because koanf
  hides which keys are live).
- Push-model struct tags make the key set a compile-time-visible fact of the type;
  discoverability becomes `Keys[T]()` — a map over the type, not a debugging session.
- Dependency floor: zero third-party deps for env+defaults+flags precedence. YAML file
  support (fleet's actual file format — dnsblockd/quotaboard both YAML) adds exactly
  `go-faster/yaml` (the how-to-golang policy yaml lib); JSON rides stdlib v2.

## 3. Tag grammar (frozen for v0.1.0)

```go
type Config struct {
    Addr     string        `env:"APP_ADDR" default:":8080"`
    LogLevel string        `env:"APP_LOG_LEVEL" default:"info"`
    APIKey   string        `env:"APP_API_KEY,required"`     // missing → Rejection
    TTL      time.Duration `env:"APP_TTL" default:"60s"`
    Hosts    []string      `env:"APP_HOSTS"`                // comma-separated
    Nested   Sub           `env:"APP"`                      // prefix for child fields
}
```

- `env:"NAME"` — binds the environment variable NAME. Absent tag = field is never
  touched by the loader (callers may pre-set it; this is what makes embedding
  appkit.ServiceConfig safe — only tagged fields bind).
- `env:"NAME,required"` — unset or empty is a `FieldError{Kind: MissingRequired}`.
- `env:"PREFIX"` on a struct field — child env keys are `PREFIX_CHILD` (underscore
  join). One level of nesting documented; deeper nesting works recursively.
- `default:"..."` — applied when env and file are both absent. Defaults are parsed
  with the SAME parser as env values, so a bad default is caught at Load time, never
  silently coerced (the koanf WeaklyTypedInput trap the F1 spec warns about).
- Supported scalar types: string, bool, all int/uint widths, float32/64,
  time.Duration, []string (comma-split, surrounding whitespace trimmed). Anything
  else tagged `env` fails Load with `Kind: UnsupportedType` — no silent skips.

## 4. Precedence (fixed, documented, tested as a matrix)

```
flag > env > file > default
```

- `default` — the tag.
- `file` — `WithFile(path)`, `.yaml`/`.yml` (go-faster/yaml) or `.json` (stdlib v2);
  keys match field names case-insensitively (the YAML fleet writes lowercase keys).
- `env` — `LookupEnv` over the process environment (injectable via `WithLookup` for
  tests — never mutate global env in library tests).
- `flag` — cmdguard stays the flag engine; the preset app calls
  `WithValues(map[string]string{"Addr": flagAddr})` (field name → raw string) after
  flag parse. Highest precedence, last write wins.

Each layer only writes fields it has a value for; a later layer overriding an earlier
one re-parses from its own raw string (no cross-layer type guessing).

## 5. Errors — Rejection + typed field path

```go
type FieldError struct {
    Field string  // struct path, e.g. "Nested.Addr"
    Key   string  // the env key or file key that failed
    Kind  FieldErrorKind // MissingRequired | ParseFailed | UnsupportedType | FileRead | FileDecode | UnknownExtension
    Err   error
}
```

`Load` returns nil cfg + a single go-error-family **Rejection** (`config.invalid`)
whose message joins every FieldError (`Field=Key: reason`), and which `errors.As`
unwraps to `[]FieldError` via the joined chain. Consumers assert `errors.As` for the
programmatic path; the string form is human-friendly for fatal(exit) output.

## 6. API surface (v0.1.0)

```go
func Load[T any](opts ...Option) (*T, error)          // one-shot
type Loader[T any] struct{ ... }                      // reusable (re-Load on SIGHUP later)
func NewLoader[T any](opts ...Option) *Loader[T]
func (l *Loader[T]) Load() (*T, error)

type Option func(*loaderConfig)
func WithFile(path string) Option                     // file layer (optional)
func WithLookup(fn func(string) (string, bool)) Option// env source, default os.LookupEnv
func WithValues(vals map[string]string) Option        // flag layer: field name → raw value
func WithValidate(fn func(*T) error) Option           // cross-field check, runs last
func Keys[T any]() []string                           // every env key the type declares (sorted)
```

Non-goals (documented in README): hot-reload/watch (F4), interpolation/substitution,
remote sources, multi-file merge, dotenv loading (compose in main if wanted).

## 7. Acceptance (from the plan, T2.2a/b)

- Table tests: precedence matrix (flag>env>file>default), missing-required →
  Rejection + FieldError path, bad-value → classified error never panic, defaults
  applied when absent, nested prefixes, empty-string env counts as unset.
- README snippets compile-checked in a scratch module (the doc-snippets-are-code rule).
- Module verify loop green hermetic (`GOWORK=off GOTOOLCHAIN=go1.27.2 go test ./... -race -count=1`).
