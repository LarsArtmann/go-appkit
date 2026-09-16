package health_test

import (
	"context"
	"errors"
	"testing"

	"github.com/larsartmann/go-appkit/health"
	gohealth "github.com/larsartmann/go-health"
)

// checkFailed reports whether a check result carries a failure message
// (Check.Error is the failure text; empty means the check passed).
func checkFailed(check gohealth.Check) bool {
	return check.Error != ""
}

// FuzzNewProbe_PanicIsolation pins the panic-isolation contract: a check
// that panics with ANY value fails as THAT CHECK's error; the batch still
// completes and healthy checks report nil. The fuzzer hunts for a panic
// value that escapes isolation (which would poison the whole batch).
func FuzzNewProbe_PanicIsolation(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(1))
	f.Add(int64(2))

	f.Fuzz(func(t *testing.T, seed int64) {
		panickingName := "panicking"

		checks := map[string]health.CheckFunc{
			panickingName: func(context.Context) error {
				switch seed % 3 {
				case 0:
					panic("fuzz-string-panic")
				case 1:
					panic(errors.New("fuzz-error-panic"))
				default:
					panic(nil) //nolint:nilpanic // deliberate: panicking with nil is the nastiest case
				}
			},
			"healthy": func(context.Context) error { return nil },
		}

		probe := health.NewProbe(checks)

		// Evaluate runs the batch synchronously and returns the full
		// snapshot; the panic-isolation contract lives in the check wrapper.
		resp := probe.Evaluate(context.Background())

		if resp.Status == "" {
			t.Fatal("empty response status after batch")
		}

		healthy, healthyOK := resp.Checks["healthy"]
		if !healthyOK || checkFailed(healthy) {
			t.Errorf("healthy check must still pass: results=%v", resp.Checks)
		}

		panicked, panickingOK := resp.Checks[panickingName]
		if !panickingOK || !checkFailed(panicked) {
			t.Errorf("panicking check must fail with an error, got ok=%v check=%+v", panickingOK, panicked)
		}
	})
}
