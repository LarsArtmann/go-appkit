# Research Reports Index

Deep-dive reports on the fleet's third-party library adoption. Scored audits follow the
`library-deep-dive` methodology (prior-report overlap → usage catalog → live capability
research → gap analysis → weighted score). Newer reports on the SAME library supersede
older ones; supersede notes below make the chain explicit.

| Report | Library / scope | Score | Supersedes / superseded |
| --- | --- | --- | --- |
| [2026-10-18 · OTel deep dive](2026-10-18_otel-deep-dive.html) | `go.opentelemetry.io/otel` v1.47.0 + `otelhttp` v0.72.0 across the repo | **78/100** (high adoption; 25/37 leveraged, 4 missed, 0 anti-patterns) | The 2026-10-11 audit's otel rows and the flight-recorder report's "bridge wired to nothing" finding (superseded by otel v0.2.0) |
| [2026-10-11 · Sibling library utilization audit](2026-10-11_sibling-library-utilization-audit.html) | Cross-cutting: the sibling-library fleet (httputil, go-error-family, go-health, go-sse, ...) | audit (utilization verdicts per library) | — |
| [2026-10-06 · go-cqrs-lite system/metaengine deep dive](2026-10-06_go-cqrs-lite-system-metaengine-deep-dive.html) | `go-cqrs-lite` `system` composition root + `metaengine` planner | deep dive | — |
| [2026-10-06 · go-flightrecorder deep dive](2026-10-06_go-flightrecorder-deep-dive.html) | `go-flightrecorder` (+ the appkit bridges) | deep dive | Its otel-bridge finding superseded 2026-10-18 (otel v0.2.0 wired the bridge) |

House rules: reports are point-in-time snapshots — annotate, never rewrite (dated errata
blocks); execution state lives in TODO_LIST/ROADMAP, not here; the 2026-10-18 report's
§f queue is the HARVEST feed for the TODO_LIST.
