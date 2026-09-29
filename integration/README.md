# integration

Cross-module and cross-repo E2E composition tests. NEVER RELEASED — it
pins the LATEST PUBLISHED tag of every family module (pin contract in
`doc.go`, values checked by `pin_drift_test.go`, guarded by
`scripts/check-pin-drift.sh`) so it always tests what consumers
resolve.

## Build & verify

```bash
cd integration
GOWORK=off GOTOOLCHAIN=go1.27.1 go mod tidy
GOWORK=off GOTOOLCHAIN=go1.27.1 go test ./... -race -count=1
GOWORK=off GOTOOLCHAIN=go1.27.1 go vet ./... && GOWORK=off GOTOOLCHAIN=go1.27.1 golangci-lint run ./...
```

Read the PINNED module's API (module cache or `git show <tag>:<file>`)
before writing tests against it — the working tree may carry unreleased
APIs the pin charter will (correctly) refuse to compile.
