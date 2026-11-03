# Reviews

Point-in-time review artifacts (scans, audits, deep-dives). Newest first.

| Date       | Artifact                                  | Scope                                                                 |
| ---------- | ----------------------------------------- | --------------------------------------------------------------------- |
| 2026-11-03 | `../review/libs/index.html` (+ 12 per-lib reports in `../review/libs/`) | Sibling-library fleet audit: httputil 72/100, go-sse 81/100, go-error-family 46/100; 9 deliberate non-uses re-verified with demand-gated adoption paths |
| 2026-10-12 | `2026-10-12_13-55_code-quality-scan.html` | Full-repo quality scan (build, lint, duplication) at the 1.27.2 floor |

Per-library utilization reports live canonically in `../review/libs/` (one md
per lib + `index.html`); this index tracks repo-level reviews.

Add one row per new review; historical status reports live in `docs/status/`,
plans in `docs/planning/`.
