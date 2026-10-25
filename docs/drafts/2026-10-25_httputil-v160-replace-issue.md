> [!NOTE]
> This filing was drafted by GLM-5.3-flash via Crush from an AI-run investigation during the go-appkit review session and filed at my request.
>
> - [x] MANUALLY REVIEWED by `@Lars Artmann` at `[2026-10-25 17:32 CET]`

**TL;DR:** The v1.6.0 go.mod carries `replace github.com/larsartmann/httputil/server_timing => ./server_timing`, and the line is still on master — so anything that loads httputil as the MAIN module refuses it (`go doc @v1.6.0` errors), and the tag breaks the family's own no-replace release rule. Consumers build fine (replace is ignored in dependency position — verified). Ask: drop the line on master, tag v1.6.1.

## Problem

v1.6.0 as served by proxy.golang.org (sumdb-verified `.mod`):

```
replace github.com/larsartmann/httputil/server_timing => ./server_timing
```

Still present in `go.mod` on master (checked 2026-10-25). With Go 1.27.2:

```
$ GOWORK=off go doc github.com/larsartmann/httputil@v1.6.0 Language
doc: github.com/larsartmann/httputil@v1.6.0 (in github.com/larsartmann/httputil@v1.6.0):
	The go.mod file for the module providing named packages contains one or
	more replace directives. It must not contain directives that would cause
	it to be interpreted differently than if it were the main module.
```

## Verified impact

- **Consumers: unaffected at build time.** A scratch module (`go 1.27.2`) requiring v1.6.0 and importing the root package: `go mod tidy` and `go build ./...` both exit 0 — the directive is ignored in dependency position. go-appkit already re-pinned integration to v1.6.0; suites + guards green.
- **Main-module-context tooling: hard error.** `go doc @v1.6.0 <symbol>` refuses the module (output above); anything rendering or loading it as the main module hits the same refusal.
- **Fleet rule:** go-appkit's `pre-tag-checks.sh` asserts every tagged artifact carries NO filesystem `replace` — v1.6.0 fails that gate as a consumer-side fact, which is how this surfaced (pin-drift demanded the bump; the bump exposed the tag).

## Cause and the cheap fix

`server_timing` is a nested module (required as v1.0.3), so the line looks like local-dev convenience left in the commit. Removing it is consumer-neutral: MVS already resolves `github.com/larsartmann/httputil/server_timing` v1.0.3 from the proxy. For local cross-module work, an untracked `go.work` replaces the replace.

## Ask

1. Delete the replace line from master `go.mod`.
2. Tag **v1.6.1**. (If you would rather force-move v1.6.0: the stale go VCS-cache/checksum-mismatch gotcha applies to every machine that already downloaded the zip — see go-appkit AGENTS.md "Stale go VCS cache after a re-tag" for the repair recipe.)

💘 Generated with Crush

Assisted-By: Crush:glm-5.3-flash
