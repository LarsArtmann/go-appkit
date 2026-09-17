# Recipe: Fresh-consumer proxy check

**What it proves:** a pushed tag resolves from `proxy.golang.org` for a
consumer with no local workspace, no `replace`, and no cache — the exact
situation of every real user after `go get`.

**When:** the LAST step of every release ritual (AGENTS.md "Release
Ritual" step 3): after `git push --tags`, before the train is closed.

**The contract is the PROXY, not pkg.go.dev.** A 404 on the pkg.go.dev
website is crawler lag (minutes-to-hours) and is NOT a release failure;
`go get` succeeding against the proxy is. Render checks are a separate,
slower ritual (TODO_LIST pkg.go.dev item).

## Fast path (script, generic for any module)

```bash
go run doc/recipes/_proxycheck/main.go \
    github.com/larsartmann/go-appkit@v0.5.1 \
    github.com/larsartmann/go-appkit/realtime@v0.1.1
```

Exit 0 = every target fetched and built as a fresh consumer. The script
creates a throwaway module per target, `go get`s the exact version,
blank-imports it, and runs `go build` with `GOWORK=off`.

## Full path (manual, includes a behavioral check for core)

1. **Scratch module outside the repo** (no workspace interference):

   ```bash
   SCRATCH=$(mktemp -d /tmp/proxycheck-XXXX) && cd "$SCRATCH"
   printf 'module proxycheck\n\ngo 1.26\n' > go.mod
   ```

2. **Fetch the exact tag** — never `@latest` (you would not be testing
   the tag you just pushed):

   ```bash
   GOWORK=off go get github.com/larsartmann/go-appkit@v0.5.1
   ```

3. **Blank import + build** (compile-level proof):

   ```go
   // main.go
   package main

   import (
       "context"
       "log"
       "net/http"
       "os/signal"
       "syscall"
       "time"

       "github.com/larsartmann/go-appkit"
   )

   func main() {
       cfg := appkit.DefaultServiceConfig()
       cfg.Addr = "127.0.0.1:0" // EPHEMERAL, always — port 8080 is SigNoz on dev boxes
       svc, err := appkit.NewService(cfg)
       if err != nil {
           log.Fatal(err)
       }
       svc.Mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
           w.WriteHeader(http.StatusOK)
       })
       errCh, err := svc.Start()
       if err != nil {
           log.Fatal(err)
       }
       log.Printf("listening on %s", svc.Addr())

       // Start does NOT install signal handling — that is Run's job. A
       // probe binary therefore handles SIGTERM itself to exercise the
       // graceful path (without this, SIGTERM hard-kills: no drain, no
       // shutdown phase logs, errCh never delivers).
       ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
       defer stop()

       select {
       case <-ctx.Done():
           shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
           defer cancel()
           if err := svc.Shutdown(shutdownCtx); err != nil {
               log.Fatal(err)
           }
       case serveErr := <-errCh:
           if serveErr != nil {
               log.Fatal(serveErr)
           }
       }
   }
   ```

   ```bash
   GOWORK=off go build ./... && GOWORK=off go vet ./...
   ```

4. **Behavioral probe (Go, never curl):** run the binary above in the
   background, then from a second scratch file (or `go run`) issue
   `http.Get("http://" + addr + "/health/ready")` (core registers
   `/health`, `/health/live`, `/health/ready` by default) and one
   authenticated `/metrics` request when the tag carries Metrics config;
   expect 200s and the `appkit_build_info` metric line. Send SIGTERM to
   the process and confirm the shutdown phase log lines end with
   `result=ok` (the default DrainDelay adds a 5 s wait — fine in a
   consumer check; pass `DrainDelay: appkit.NoDrainDelay` to skip it).

## Failure modes this ritual has actually hit

| Symptom | Cause | Fix |
| ------- | ----- | --- |
| `missing <module>/go.mod at revision <tag>` | ghost tag: module path ≠ tagged directory (docs v0.2.0) | fix the layout, re-tag with a NEW version — never re-push a tag |
| fetches sibling workspace code / checksum surprises | `go.work` interference | always `GOWORK=off` for every go command in the scratch module |
| `panic: service did not start` during the behavioral probe | fixed port already owned (8080 = SigNoz here) | `Addr: "127.0.0.1:0"` in the consumer, read the returned addr — this has bitten THREE sessions |
| `go get` works but the website 404s | pkg.go.dev crawler lag | not a failure; re-check the render later (TODO_LIST owns the render item) |
| `cannot find module ...@latest`-style confusion | used `@latest` instead of the explicit tag | pin the exact `module@version` under test |

## History

Hand-rolled three times (2026-09-04 waves, 2026-09-16 train, 2026-09-17
core v0.5.0) before being committed as this recipe + script (2026-09-17,
SUPERB plan v3 task C4). Submodule pages need module-root LICENSE files
to index at all (landed 2026-09-04).
