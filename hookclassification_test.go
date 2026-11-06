package appkit

import (
	"context"
	"errors"
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-error-family/errorfamilytest"
)

// errForeignHookFailure is the unclassified control error: runHooks must
// wrap it with the phase code (unlike classified hook errors, which keep
// their own family).
var errForeignHookFailure = errors.New("plain failure")

// multiUnwrapper is the shape errors.Join returns: an error that exposes its
// members. errors.AsType constrains its type argument to error, so the
// anonymous `interface{ Unwrap() []error }` needs the Error method folded in.
type multiUnwrapper interface {
	error

	Unwrap() []error
}

// TestRunHooks_RespectsHookClassification pins the WrapOnce contract
// (error-family adoption batch, audit F3): a hook that already returns an
// errorfamily error keeps its own family and code — no re-wrap nesting under
// the phase code — while a foreign error gains the phase classification;
// both carry their hook position in the error context.
func TestRunHooks_RespectsHookClassification(t *testing.T) {
	t.Parallel()

	classified := errorfamily.NewRejection("app.my_hook_rejected", "hook knows better")

	hooks := []Hook{
		func(context.Context) error { return nil }, // index 0: silent
		func(context.Context) error { return classified },
		func(context.Context) error { return errForeignHookFailure },
	}

	svc := &Service{}

	joined := svc.runHooks(
		context.Background(), hooks, "server.shutdown_hook_failed", "shutdown hook failed",
	)

	joinedSplit, ok := errors.AsType[multiUnwrapper](joined)
	if !ok {
		t.Fatalf("joined error does not unwrap to members: %v", joined)
	}

	members := joinedSplit.Unwrap()
	if len(members) != 2 {
		t.Fatalf("joined error splits into %d members, want 2 (silent hook contributes none): %v", len(members), joined)
	}

	errorfamilytest.AssertFamily(t, members[0], errorfamily.Rejection)
	errorfamilytest.AssertCode(t, members[0], "app.my_hook_rejected")
	errorfamilytest.AssertContext(t, members[0], "hook_index", "1")

	errorfamilytest.AssertFamily(t, members[1], errorfamily.Infrastructure)
	errorfamilytest.AssertCode(t, members[1], "server.shutdown_hook_failed")
	errorfamilytest.AssertContext(t, members[1], "hook_index", "2")
}
