package appkit

import (
	"log/slog"

	"github.com/larsartmann/httputil"
	servertiming "github.com/larsartmann/httputil/server_timing"
)

// defaultMiddlewareStack returns the opinionated middleware chain:
// ServerTiming → Recovery → RequestID → Logging → Timeout → SecurityHeaders.
// ServerTiming (W3C Server-Timing, the family's httputil/server_timing
// module) sits outermost so its auto-recorded total covers the entire
// downstream chain, recovery included; the header is stamped when the
// response commits, so SSE streams (WriteTimeout = NoTimeout) are
// unaffected. Handlers record sub-phases via the request-context collector
// BEFORE their first write — post-commit metrics miss the header (the
// library's stamp-on-commit contract), which is why the framework records
// no phase of its own:
//
//	stop := servertiming.MeasureServerTiming(r.Context(), "db")
//	rows, err := db.QueryContext(r.Context(), ...)
//	stop()
//
// With ServerTiming: &false the middleware is omitted entirely. With
// WriteTimeout = NoTimeout the Timeout middleware is omitted: an
// already-expired or zero context deadline would cut long-lived streams
// that the operator explicitly asked to run unbounded.
func defaultMiddlewareStack(logger *slog.Logger, cfg ServiceConfig) []httputil.Middleware {
	var stack []httputil.Middleware

	if cfg.ServerTiming == nil || *cfg.ServerTiming {
		stack = append(stack, servertiming.ServerTimingMiddleware())
	}

	stack = append(stack,
		httputil.Recovery(logger),
		httputil.RequestID(httputil.DefaultRequestIDConfig()),
		httputil.Logging(logger),
	)

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
