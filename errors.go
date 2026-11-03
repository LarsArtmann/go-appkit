package appkit

import (
	"log/slog"
	"net/http"

	errorfamily "github.com/larsartmann/go-error-family"
)

// RegisterClassifications installs appkit's fleet-default error
// classifications into the go-error-family default registry, so errors that
// cross appkit boundaries classify correctly at Classify/HTTPStatus
// boundaries without per-module ad-hoc mapping:
//
//   - the stdlib defaults (context deadline/cancellation, sql.ErrNoRows, os
//     sentinels — go-error-family carries the rationale table);
//   - *http.MaxBytesError — the typed contract of the MaxBytesReader wrapper
//     (security.BodyLimit, httputil.MaxBodySize) — classifies as Rejection,
//     so an oversize body maps to the 400 family instead of the fail-open
//     Transient default.
//
// Call once at boot, before serving. Re-registration is idempotent (same
// keys overwrite). Deliberately NOT called on import or by NewService:
// classification registries are process-global state, and a library must
// not mutate global state as an import side effect. Consumers preferring
// scoped registries can build their own (NewRegistry + Clone) and register
// the same types there.
func RegisterClassifications() {
	errorfamily.RegisterStdlibDefaults(errorfamily.DefaultRegistry)
	errorfamily.RegisterClassificationType[*http.MaxBytesError](errorfamily.Rejection)
}

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
