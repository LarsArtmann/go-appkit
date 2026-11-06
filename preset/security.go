package preset

import (
	"log/slog"

	"github.com/larsartmann/go-appkit/security"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/httputil"
)

// securityMiddlewares composes the profile from shipped security v0.2.0
// batteries only — the preset adds no capability of its own.
//
//	standard: SecurityHeaders(Production) + BodyLimit(1 MiB default)
//	strict:   standard + OriginCheck(Origins) — Origins REQUIRED
func securityMiddlewares(cfg *SecurityConfig) ([]httputil.Middleware, error) {
	profile := SecurityStandard
	bodyLimit := defaultBodyLimitBytes

	var origins []string

	if cfg != nil {
		if cfg.Profile != "" {
			profile = cfg.Profile
		}

		if cfg.BodyLimitBytes > 0 {
			bodyLimit = cfg.BodyLimitBytes
		}

		origins = cfg.Origins
	}

	if profile == SecurityStrict && len(origins) == 0 {
		return nil, errorfamily.NewRejection(
			"preset.strict_origins_missing",
			"strict security profile requires Config.Security.Origins (an empty allowlist would reject every browser request)",
		)
	}

	middlewares := []httputil.Middleware{
		security.SecurityHeaders(
			security.HeadersConfig{ //nolint:exhaustruct_v5 // Production tier carries production defaults for the rest
				Environment: security.Production,
			},
		),
		security.BodyLimit(bodyLimit),
	}

	if profile == SecurityStrict {
		middlewares = append(middlewares, security.OriginCheck(origins, slog.Default()))
	}

	return middlewares, nil
}
