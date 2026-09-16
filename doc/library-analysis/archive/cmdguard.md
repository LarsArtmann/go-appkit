# cmdguard — Integration Analysis

> **Verdict: 🟩 NOT APPLICABLE (wrong layer — CLI, not server library).** cmdguard is a type-safe
> Cobra wrapper for building command-line interfaces. go-appkit is a server skeleton with **no
> command surface**. CLI concerns belong in the consuming service's `main` package, not in an
> imported library.

---

## 1. Library identity

| Attribute      | Value                                                                       |
| -------------- | --------------------------------------------------------------------------- |
| Module path    | `github.com/larsartmann/cmdguard/v2`                                        |
| Public import  | `github.com/larsartmann/cmdguard/v2/pkg/cmdguard/v2`                        |
| Go version     | 1.26.4                                                                      |
| Latest release | **v2.10.2** (v1 is legacy/unsupported)                                      |
| License        | MIT                                                                         |
| Status         | v2 stable (additive-only until v3)                                          |
| Sibling deps   | `go-output`, `samber-do-auditlog`; indirect `go-branded-id` (via go-output) |

A type-safe wrapper over **Cobra** that makes correct Cobra behavior the default: `SilenceUsage`
on, single error print, correct exit codes, only error-returning handlers (no panics),
construct-time validation, struct-tag-driven typed flags, DI via `samber/do/v2`, styled output via
`fang`.

---

## 2. What it provides (for completeness)

- **`CLI[T]`** — generic orchestrator over config type `T`. `NewCLI[T](name, short, defaults, opts…)`.
  Methods: `Execute`, `ExecuteAndExit`, `ExecuteWithArgs`, `Scope`, `Config`, `Shutdown`,
  `HealthCheck`, `GenerateDocs`, `ManPage`, …
- **`Command[T, F]`** / **`NewCommand`** / **`NewParentCommand`** / **`AddCommand`** — each command
  carries its own typed flag struct `F`; handler signature `func(ctx, *T, F) error`.
- **`Scope`** — DI over `samber/do/v2` (`Provide`/`Invoke`/`Override`/`Shutdown`).
- **`FlagRegistry`** + struct tags: `flag`, `short`, `default`, `help`, `validate`, `env`, `prompt`,
  `values`, `required`, `count`, `local`, `hidden`. Priority: flag → env → config file → default.
- Built-in value types: `Duration`, `Enum`, `LogLevel`, `LogFormat`, `URL`, `Email`, `Port`,
  `FilePath`, `HostPort`.
- `Result[T]` / `Validated[T]` sum types; `Middleware[T]`; `Plugin` interface; `configload/`
  (YAML/TOML/Auto/Koanf); 60+ sentinel errors; `ExitCode(err)` mapper.
- Zero panics (all `Must*` removed in v2.5.0).

---

## 3. Current usage in go-appkit

**Zero.** Not in `go.mod`; no imports. appkit has no `main`, no commands, no flags, no CLI surface —
it is a library (`package appkit`) consumed by services.

---

## 4. Applicability assessment — why "wrong layer"

CLI frameworks live in the **`main` package of an executable**, because:

1. Commands, flags, and `os.Args` parsing are the entrypoint's job — there's exactly one `main` per
   binary.
2. A library that imports a CLI framework drags Cobra, pflag, fang, huh, glamour, lipgloss, bubbles,
   otel, koanf… into **every importer**, including services that are not CLIs (HTTP servers,
   workers, daemons). appkit's consumers are overwhelmingly servers — they'd pay the cost for nothing.
3. appkit's own surface (`NewServer`, `OpenSQLite`, `InitLogger`, `WaitForSignal`) is **programmatic
   Go API**, not CLI commands. There is nothing to wrap in Cobra.

| Hypothetical use in appkit                    | Reality                                                                                                                          |
| --------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| Wrap appkit setup behind a `cmdguard.CLI`?    | That's an application's `main`, not a library.                                                                                   |
| Offer a `appkit serve` command?               | appkit is imported, not run. Services define their own commands.                                                                 |
| Use `configload` for appkit's `LoggerConfig`? | appkit takes structs as args; file loading is the application's concern.                                                         |
| Use `LogLevel`/`LogFormat` value types?       | appkit already defines its own `LogLevel`/`LogFormat` typed strings — and importing cmdguard just for two enums would be absurd. |

---

## 5. Integration analysis — if someone mistakenly tried

- **Cost:** severe. cmdguard's dependency tree (Cobra, pflag, charm.land suite, otel, koanf, samber/do,
  go-output + 9 sub-modules) is far heavier than appkit's entire current footprint. appkit would
  cease to be a "tiny skeleton."
- **Correctness:** a library importing a CLI framework is a layering smell. Libraries expose APIs;
  executables define commands.
- **Direction:** not inverted per se, but inappropriate — cmdguard is a tool for `main` packages.

---

## 6. What "fully" would mean here

Nothing coherent. "appkit uses cmdguard fully" would mean appkit becomes a CLI framework, which
contradicts its purpose as an importable server skeleton.

---

## 7. Recommendation

- **Do not add cmdguard to appkit.** CLI construction belongs in each service's `main` package.
- If a _service_ built on appkit wants a great CLI, that service should use cmdguard directly in its
  `main` — calling `appkit.NewServer` / `appkit.OpenSQLite` from inside a `cmdguard.Command` handler.
  That is the natural composition; no appkit change needed.
- Note an existing echo: appkit already defines `LogLevel`/`LogFormat`, and cmdguard ships built-in
  `LogLevel`/`LogFormat` value types. This is a mild duplication _between projects_, but resolving it
  by importing cmdguard into appkit would be worse than the duplication. Leave as-is.

---

## 8. Summary

- **Using today?** No.
- **Fully?** No — not applicable. appkit is a library with no command surface; CLI concerns belong in
  `main`.
- **Action:** None. Services that want a CLI compose cmdguard in their own `main`, calling appkit's
  programmatic API from command handlers.
