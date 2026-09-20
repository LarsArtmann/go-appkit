# Upstream draft: samber/do — document that UNBUILT lazy services report healthy

**Status:** DRAFT — filing is USER-gated (plan Gate Q3, 2026-09-20). Do not
file without explicit approval.
**Target repo:** github.com/samber/do v2.1.0 (EXTERNAL repo — full
verify-before-filing gates applied, see self-review below)
**Where:** documentation gap (README / `Healthchecker*` doc comments); the
behavior itself is presumably intended lazy semantics and is NOT challenged.

## The ask

Document — in the README's health-check section (if one exists; today
`Healthchecker`/`HealthcheckerWithContext` appear only in code) and in the
`Healthchecker*` interface doc comments — that for a **lazy** service which
has never been invoked, the injector's health check reports **healthy**
(nil error) regardless of the real dependency state.

## Evidence (source, v2.1.0)

`service_lazy.go:128`:

```go
func (s *serviceLazy[T]) healthcheck(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.built {
		return nil // <- unbuilt lazy service: healthy, unconditionally
	}
	// ... delegate to the instance's HealthCheck[WithContext]
}
```

Nothing in the README mentions health checks at all (grep for
"healthcheck" over v2.1.0's README returns no hits), so a user wiring
`HealthcheckerWithContext` onto a lazy service has no way to learn this.

## Real-world impact (why it matters beyond pedantry)

Health-check ABSTRACTIONS built on do inherit the silent pass. Concretely:
a bridge library (go-appkit/flightrecorderhealth) that registers consumers'
dependencies for a go-health probe must eagerly `InvokeNamed` every
registered service — if it forgets one, or a consumer registers its own
lazy health-checkable service, that dependency reports healthy until its
first unrelated resolution. A failing database can therefore sit behind a
green readiness page. The bridge now does the eager invoke and documents it
as load-bearing — but the semantics causing the ceremony are invisible in
do's own docs.

## Proposed change (minimal)

Two sentences wherever health checks are (or become) documented, e.g.:

> A lazy service that has not been built yet reports healthy (its
> `Healthchecker*` method has not run). Call `MustInvoke`/`Invoke` once at
> startup — or use `ProvideValue` — if health checks must reflect real
> dependency state from the first check.

Optionally, a `do.MustInvokeNamed` callout in the health-check context
would give readers the remedy next to the trap.

## Self-review (verify-before-filing checklist, 2026-09-20)

- Source read, not guessed: `service_lazy.go:128-134` (v2.1.0, pinned in the
  module cache).
- Not already documented: README has zero healthcheck mentions; the
  interface doc comments do not mention built-state either.
- Would the proposed change fix it: yes — it is a documentation gap; the
  remedy (eager invoke) is already idiomatic do.
- Behavior not challenged: lazy-until-invoke is do's core contract; the ask
  is only that its health-check interaction be written down.
- Existing alternatives checked: do's `ProvideValue`/eager services exist
  and are referenced as the remedy; no existing README section contradicts.
- Annoyance test: two-sentence docs addition with a concrete failure story —
  low cost, real consumer impact.
