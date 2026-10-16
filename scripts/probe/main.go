// Command probe is the example-smoke harness's HTTP prober: fetch a URL,
// exit 0 when the status code matches the expectation, 1 otherwise. Kept
// dependency-free and output-free so scripts/example-smoke.sh can use it
// from any module context.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	wantArgs     = 3
	probeTimeout = 5 * time.Second
)

func main() {
	if len(os.Args) != wantArgs {
		fmt.Fprintln(os.Stderr, "usage: probe <url> <want-status-code>")

		os.Exit(2)
	}

	url, want := os.Args[1], os.Args[2]

	client := &http.Client{Timeout: probeTimeout}

	//nolint:gosec // one-shot CLI probe: operator-supplied URL
	res, err := client.Get(url) //nolint:noctx // one-shot CLI probe, no cancellation needed
	if err != nil {
		fmt.Fprintf(os.Stderr, "probe %s: %v\n", url, err)

		os.Exit(1)
	}

	_, _ = io.Copy(io.Discard, res.Body)
	_ = res.Body.Close()

	if strconv.Itoa(res.StatusCode) != want {
		fmt.Fprintf(os.Stderr, "probe %s: status %s, want %s\n", url, res.Status, want)

		os.Exit(1)
	}
}
