package appkit_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/larsartmann/go-appkit"
	errorfamily "github.com/larsartmann/go-error-family"
)

func TestRegisterClassifications_MaxBytesErrorIsRejection(t *testing.T) {
	t.Parallel()

	// Idempotent: boot code may call it defensively more than once.
	appkit.RegisterClassifications()
	appkit.RegisterClassifications()

	limitErr := &http.MaxBytesError{Limit: 1 << 20}

	if got := errorfamily.Classify(limitErr); got != errorfamily.Rejection {
		t.Errorf("Classify(*http.MaxBytesError) = %v, want Rejection", got)
	}

	if got := appkit.HTTPStatus(limitErr); got != http.StatusBadRequest {
		t.Errorf("HTTPStatus(*http.MaxBytesError) = %d, want %d", got, http.StatusBadRequest)
	}
}

func TestRegisterClassifications_StdlibDefaults(t *testing.T) {
	t.Parallel()

	appkit.RegisterClassifications()

	if got := errorfamily.Classify(context.DeadlineExceeded); got != errorfamily.Transient {
		t.Errorf("Classify(context.DeadlineExceeded) = %v, want Transient (stdlib default)", got)
	}

	if got := errorfamily.Classify(context.Canceled); got != errorfamily.Rejection {
		t.Errorf("Classify(context.Canceled) = %v, want Rejection (stdlib default: caller abandoned)", got)
	}
}
