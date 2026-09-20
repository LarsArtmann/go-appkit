package flightrecorderhealth

import (
	"errors"
	"fmt"
	"testing"
)

// failingServiceNames feeds trigger log lines; sorted output keeps them
// diff-able and test assertions deterministic despite randomized map order.
// Run repeatedly to make map-order flakiness probabilistically visible.
func TestFailingServiceNames_SortedDeterministic(t *testing.T) {
	t.Parallel()

	results := map[string]error{
		"zebra":   errors.New("down"),
		"alpha":   errors.New("down"),
		"healthy": nil,
		"mid":     errors.New("down"),
	}

	want := "[alpha mid zebra]"
	for i := range 50 {
		got := fmt.Sprint(failingServiceNames(results))
		if got != want {
			t.Fatalf("iteration %d: failingServiceNames = %s, want %s", i, got, want)
		}
	}
}

func TestFirstError_ReturnsANonNilError(t *testing.T) {
	t.Parallel()

	errDB := errors.New("db down")
	errCache := errors.New("cache down")
	results := map[string]error{
		"database": errDB,
		"cache":    errCache,
		"healthy":  nil,
	}

	// WHICH failure surfaces is unspecified (randomized map order) — the
	// contract is "a representative non-nil failure", documented on
	// firstError. Assert the stable part of the contract.
	for range 50 {
		got := firstError(results)
		if got == nil {
			t.Fatal("firstError = nil, want a representative failure")
		}
	}

	representative := firstError(map[string]error{"ok": nil})
	if representative != nil {
		t.Fatalf("firstError(all-pass) = %v, want nil", representative)
	}
}
