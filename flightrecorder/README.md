# flightrecorder

HTTP middleware + snapshot endpoint over
[go-flightrecorder](https://github.com/larsartmann/go-flightrecorder):
capture Go runtime traces on configurable triggers (error threshold) or
on demand. Package docs live in `doc.go`.

## Build & verify

```bash
cd flightrecorder
GOWORK=off GOTOOLCHAIN=go1.27.1 go test ./... -race -count=1
GOWORK=off GOTOOLCHAIN=go1.27.1 go vet ./... && GOWORK=off GOTOOLCHAIN=go1.27.1 go build ./...
golangci-lint run ./...        # from this directory
```

`GOTOOLCHAIN=go1.27.1` is required (imports `encoding/json/v2`
directly). Tests serialize on `recorderMu`: Go's `runtime/trace` allows
one active flight recorder per process. Use `fr.WithWriter` (not
`fr.WithFile` + delete) for multi-capture tests — file handles are
cached on first write.
