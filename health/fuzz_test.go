package health_test

import (
	"context"
	"errors"
	"testing"

	"github.com/larsartmann/go-appkit/health"
)

// FuzzNewProbe_PanicIsolation pins the panic-isolation contract: a check
// that panics with ANY value fails as THAT CHECK's error; the batch still
// completes and healthy checks report nil. The fuzzer hunts for a panic
// value that escapes isolation (which would poison the whole batch).
func FuzzNewProbe_PanicIsolation(f *testing.F) {
	f.Add("string panic")
	f.Add(42)
	f.Add(nil) //nolint:nilnil // a nil panic() value is exactly the edge case

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

		results := probe.Check(context.Background()) //nolint:contextcheck // probe owns its batch context

		healthyErr, healthyOK := results["healthy"]
		if !healthyOK || healthyErr != nil {
			t.Errorf("healthy check must still pass: results=%v", results)
		}

		if err, ok := results[panickingName]; !ok || err == nil {
			t.Errorf("panicking check must fail with an error, got ok=%v err=%v", ok, err)
		}
	})
}
