# Battery-Swap Recipe: <APP>

> Copy this template per app. A recipe is a CHECKLIST executed in the CONSUMER repo —
> appkit ships the recipe; the consumer repo lands the swap. Keep every claim mechanical:
> numbers come from commands, verdicts cite files.

**App:** `<name>` (`/home/lars/projects/<name>`) · **Recipe date:** YYYY-MM-DD ·
**App suite command:** `<the one command that proves the app still works>`
**Baseline LOC to delete:** `<n>` (from §1) · **Swapped LOC after:** `<n>` (from §4)

## 1. Inventory (what the app hand-rolls today)

Per hand-rolled block: file(s), LOC (`wc -l`), and the one-line job it does.
Commands, not estimates — `wc -l internal/otel/*.go` etc.

| Hand-roll | Files | LOC | Job |
| --------- | ----- | --- | --- |

## 2. Battery map (hand-roll → shipped battery)

Each row names the appkit battery that REPLACES the hand-roll, or records a
NO-BATTERY verdict (then the hand-roll stays — an honest recipe keeps it).

| Hand-roll | Battery | Wiring | Confidence |
| --------- | ------- | ------ | ---------- |

## 3. Swap order (risk-ranked, one commit each)

Order by blast radius: observability first (otel — additive), then security
(headers/body-limit — wire-contract), then lifecycle (systemd/config — boot
path), transports last (SSE). Every step: swap → suite → commit.

1. …

## 4. LOC ledger

| Step | Deleted | Added | Net |
| ---- | ------- | ----- | --- |

Net MUST be negative or the recipe failed its point.

## 5. Suite green + sign-off

- [ ] App suite green after EVERY step (the command from the header)
- [ ] `scripts/count-adopters.sh` delta: the app now counts DIRECT for go-appkit
- [ ] Post-swap notes (what hurt, what the template/scaffold should absorb)
