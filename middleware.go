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
// OuterMiddlewares (if any) wrap everything else so instrumentation can
// observe the full request lifetime. If cfg.Middlewares is non-nil, it
// replaces the default stack entirely (OuterMiddlewares still wrap it);
// otherwise, the default stack is used with cfg.ExtraMiddlewares appended.
func buildMiddleware(logger *slog.Logger, cfg ServiceConfig) []httputil.Middleware {
	stack := cfg.Middlewares
	if stack == nil {
		stack = append(defaultMiddlewareStack(logger, cfg), cfg.ExtraMiddlewares...)
	}

	return concatMiddlewares(cfg.OuterMiddlewares, stack)
}

// concatMiddlewares returns a fresh outer→inner chain. A new slice is
// allocated so the config's slices are never written through a shared
// backing array.
func concatMiddlewares(outer, inner []httputil.Middleware) []httputil.Middleware {
	chain := make([]httputil.Middleware, 0, len(outer)+len(inner))
	chain = append(chain, outer...)
	chain = append(chain, inner...)

	return chain
}
