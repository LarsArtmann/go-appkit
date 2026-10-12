## TL;DR

Repos that defer `github.com/samber/lo` adoption by policy — library repos where a new require rides release trains — get a permanent stream of `stdlib2lo` suggestion warnings with no way to opt out. Ask: a migrator selection/deny knob in `sdk.Options`, surfaced as a toolsdk Option.

## Problem

`stdlib2lo` is registered for every consumer via the blank import in `pkg/allmigrators/allmigrators.go:21` (go-auto-upgrade @ HEAD). Its dependency-gate design is sound: when `lo` is not a go.mod require, rewrites are withheld and each site degrades to a suggestion warning plus one `lo-dependency-missing` finding per module (`internal/migrators/stdlib2lo/migrator.go:55-57`).

But "withheld" still reports every site, forever. There is no knob to say "this repo defers lo deliberately":

- `pkg/provider/provider.go:93-100`: the toolsdk `Options` list has exactly one entry (`blame`).
- The repair call (`sdk.MigrateFiles`, line ~181) passes no migrator selection — the migrator set "defaults to recommended" and nothing on the BuildFlow side can narrow it (`tool_options` validate against the exposed Options, so a consumer cannot invent one).

## Live numbers

go-appkit (12-module Go library fleet, lo deliberately not yet adopted): ~29 stdlib2lo findings on every BuildFlow run — cqrs/domain.go Map/Reduce/Flatten, realtime handler/test Filter/Map, security csrf/origin_check Filter/SliceToMap, health/probe SliceToMap, otel metrics_test Map, frh adapter FromPtr — with the root-module rollup re-counting them to 15 more. Adoption is queued for a dependency train; until then every run re-litigates the decision.

## Ask

- `sdk.Options.Migrators` (include/exclude, e.g. `"recommended,-stdlib2lo"`), surfaced as a toolsdk Option so BuildFlow `tool_options` can carry it per repo;
- or a migrator deny list in whatever config surface you prefer.

A findings-gate baseline could ack the noise per repo, but that is a per-repo ack list maintained to express one policy stance — the policy belongs at the migrator-selection layer.

💘 Generated with Crush
