package appkit

import (
	"log/slog"

	errorfamily "github.com/larsartmann/go-error-family"
)

// HTTPStatus returns the HTTP status code for a classified error.
// Returns 500 for unclassified errors. This is a re-export of
// errorfamily.HTTPStatus for consumers who import appkit.
func HTTPStatus(err error) int {
	return errorfamily.HTTPStatus(err)
}

// LogError logs an error with auto-determined severity based on its family:
// Transient → Warn, all others → Error. Structured fields are extracted
// from the error's context. This is a re-export of errorfamily.LogError.
func LogError(err error, logger *slog.Logger) {
	errorfamily.LogError(err, logger)
}
