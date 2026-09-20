# Upstream draft: go-health — reject `WithHealthRecorder` on the function-path constructors with a sentinel error

**Status:** DRAFT — filing is USER-gated (plan Gate Q3, 2026-09-20). Do not
file without explicit approval.
**Target repo:** github.com/larsartmann/go-health (own repo — treat as a
decision note, not an external ask)
**Evidence versions:** v0.1.3, v0.2.0, v0.3.0 (all verified identical)

## The ask

`NewWithHealthCheck` (and `NewWithDetailedCheck`) currently accept a
`WithHealthRecorder` option and then silently nil it. Make them reject the
option instead — return a probe-level error or panic with a sentinel such as
`ErrRecorderNotApplicable` — so a caller who wires trace capture through the
wrong constructor finds out at construction time instead of never.

## Evidence

Source, all three published versions (accessors.go, same lines):

```go
func NewWithHealthCheck(fn HealthCheckFunc, opts ...Option) *Probe {
	cfg := buildConfig(opts)
	cfg.recorder = nil // <- the option is accepted, then dropped

	return assemble(adaptPlainChecks(fn), cfg)
}
```

Runtime repro (ran against v0.2.0, 2026-09-20): a `HealthRecorder` spy passed
via `WithHealthRecorder` to `NewWithHealthCheck`, followed by
`probe.Evaluate(ctx)`, observes **0 batches** — no error, no warning, no way
to detect the drop from the returned `*Probe`.

The godoc documents the drop in prose ("If the options include
[WithHealthRecorder], it has no effect here"), so this is a known tradeoff,
not an undiscovered bug — but prose-only is a trap in practice: the flagship
trace-capture wiring (`flightrecorderhealth.NewTrigger` + `WithHealthRecorder`)
compiles and runs against the function path and silently captures nothing.
go-appkit's health module hit exactly this and now warns in its own godoc
(`appkithealth.NewProbe` forwards options to `NewWithHealthCheck`).

## Why the real-world path is fragile

The natural consumer pattern is to build the trigger first and pass it to
whichever probe constructor is at hand. With the injector-free constructors
the wiring is accepted and dead; with `New` it works. The only signal
today is reading two godoc pages carefully.

## Proposed change (minimal)

In `NewWithHealthCheck` / `NewWithDetailedCheck`, detect the recorder option
(e.g. have `WithHealthRecorder` mark the config, or build the config and
check `cfg.recorder != nil` BEFORE nil-ing it) and fail loudly:

```go
var ErrRecorderNotApplicable = errors.New("go-health: WithHealthRecorder is not applicable to function-path constructors; use New(injector, ...) instead")

func NewWithHealthCheck(fn HealthCheckFunc, opts ...Option) (*Probe, error) {
	cfg := buildConfig(opts)
	if cfg.recorder != nil {
		return nil, ErrRecorderNotApplicable
	}
	// ...
}
```

That changes the constructors' signatures (breaking at 0.x — acceptable with
a minor bump and migration note). A softer alternative: keep the signatures
and expose a `Probe.Recorder() HealthRecorder` accessor returning the
resolved recorder, so callers can assert non-nil in tests. The sentinel
error is preferred — the accessor still allows the silent mis-wiring in
production.

## Secondary observation (same repo, same evidence pass)

`Probe.Evaluate(ctx)` runs a full batch (including recorder callbacks) but
does NOT publish to the background cache that readiness handlers serve —
only the refresh loop's `refreshCache` publishes. A test that flips a
dependency and asserts through `/readyz` (or `CachedResponse`) will keep
seeing the stale pass until the next refresh tick even though `Evaluate`
observed the failure. Worth a godoc line on `Evaluate` ("does not update
the cache; use for one-off evaluations"), or a `EvaluateAndCache` variant.

## Self-review (verify-before-filing checklist, 2026-09-20)

- Source read, not guessed: `accessors.go:41/:61` in v0.1.3, v0.2.0, v0.3.0 —
  identical; upstream godoc acknowledges the drop in prose.
- Not already fixed: v0.3.0 (latest) behaves the same.
- Runtime-verified: spy recorder sees 0 batches through the function path
  (repro above); the injector path sees every batch (output-pinned example
  `ExampleNewProbe_recorderViaInjector` in go-appkit/health + the
  integration E2E `TestHealthStackThroughAppkitService`, trace bytes > 0).
- Would the proposed change fix it: yes — construction-time failure moves
  the feedback from "reading docs" to "running tests".
- Existing alternatives checked: prose godoc (current), accessor variant
  (proposed as softer alternative), appkit-side godoc warning (shipped —
  protects our consumers but not the general pattern).
- Own-repo note: go-health is Lars's repo, so "filing" is a self-decision;
  the draft exists so the decision is one command, not one investigation.
