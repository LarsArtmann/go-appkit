package appkit

import (
	"log/slog"

	"github.com/larsartmann/httputil"
)

// defaultMiddlewareStack returns the opinionated middleware chain:
// Recovery → RequestID → Logging → Timeout → SecurityHeaders.
// With WriteTimeout = NoTimeout (SSE services) the Timeout middleware is
// omitted: an already-expired or zero context deadline would cut long-lived
// streams that the operator explicitly asked to run unbounded.
func defaultMiddlewareStack(logger *slog.Logger, cfg ServiceConfig) []httputil.Middleware {
	stack := []httputil.Middleware{
		httputil.Recovery(logger),
		httputil.RequestID(httputil.DefaultRequestIDConfig()),
		httputil.Logging(logger),
	}

	if cfg.WriteTimeout != NoTimeout {
		stack = append(stack, httputil.Timeout(cfg.WriteTimeout))
	}

	return append(stack, httputil.SecurityHeaders(httputil.DefaultSecurityHeadersConfig()))
}

// buildMiddleware resolves the final middleware chain from config.
// If cfg.Middlewares is non-nil, it replaces the default stack entirely.
// Otherwise, the default stack is used with cfg.ExtraMiddlewares appended.
func buildMiddleware(logger *slog.Logger, cfg ServiceConfig) []httputil.Middleware {
	if cfg.Middlewares != nil {
		return cfg.Middlewares
	}

	stack := defaultMiddlewareStack(logger, cfg)

	return append(stack, cfg.ExtraMiddlewares...)
}
