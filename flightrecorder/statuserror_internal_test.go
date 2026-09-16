package flightrecorder

import (
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-error-family/errorfamilytest"
)

// TestStatusError_Classified pins the classification of the middleware's
// synthetic status errors: Infrastructure with the stable code, so consumer
// routing treats them like any other dependency failure.
func TestStatusError_Classified(t *testing.T) {
	t.Parallel()

	err := statusError(500, 500)
	errorfamilytest.AssertFamily(t, err, errorfamily.Infrastructure)
	errorfamilytest.AssertCode(t, err, "flightrecorder.http_status_error")
	errorfamilytest.AssertHTTPStatus(t, err, 503)

	if statusError(200, 500) != nil {
		t.Error("sub-threshold status must produce nil")
	}
}
