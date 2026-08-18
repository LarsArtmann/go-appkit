// Package otel provides opt-in OpenTelemetry instrumentation for HTTP
// services. Import it with the alias appkitotel to avoid colliding with
// go.opentelemetry.io/otel:
//
//	import appkitotel "github.com/larsartmann/go-appkit/otel"
//
// The package has no go-appkit core dependency: every piece works on plain
// net/http. In an appkit service, three wiring points make a fully traced,
// metric-emitting, correctly-flushed service:
//
//	provider, err := appkitotel.Setup(
//		appkitotel.WithService("myapp", "1.0.0", os.Getenv("POD_NAME")),
//		appkitotel.WithStdoutExporter(os.Stdout), // development; OTLP in production
//	)
//	if err != nil {
//		return err
//	}
//
//	cfg := appkit.DefaultServiceConfig()
//	cfg.OuterMiddlewares = []httputil.Middleware{appkitotel.Middleware()}
//	cfg.ShutdownHooks = []func(context.Context) error{provider.Shutdown}
//
// All instrumentation is opt-in and no-op until a provider is configured:
// without [Setup], [Middleware] propagates nothing, records nothing, and
// adds near-zero overhead — the same posture as go-cqrs-lite's otel module.
//
// # What the middleware emits
//
// [Middleware] bridges go.opentelemetry.io/contrib's otelhttp:
//
//   - One SERVER span per request, named after the matched ServeMux pattern
//     ("GET /users/{id}") when Go 1.22+ pattern routing is in play, falling
//     back to the HTTP method. Panic-safe and SSE-safe: the span ends when
//     the response stream ends.
//   - W3C trace-context + baggage propagation in both directions: incoming
//     traceparent continues the caller's trace, and outgoing client requests
//     made with otelhttp.Transport continue this service's trace.
//   - Semantic-convention metrics (http.server.request.duration and friends)
//     when a meter provider is registered, with http.route from the matched
//     pattern — cardinality-safe for parametrized routes.
//   - Health endpoints (/health, /health/live, /health/ready) are filtered
//     out by default; extend with [WithFilter] or [WithFilteredPaths].
//
// # Shutdown ordering
//
// ServiceConfig.ShutdownHooks run after the server released its connections,
// so a flushed [Provider] captures spans from the final in-flight requests.
// The listener is gone before providers shut down — the ordering telemetry
// needs.
//
// # Log correlation
//
// Wrap your slog handler with [TraceHandler] to stamp trace_id and span_id
// onto every record logged with a request context:
//
//	logger := slog.New(appkitotel.TraceHandler(slog.NewJSONHandler(os.Stdout, nil)))
//
// Handlers should log via logger.InfoContext(r.Context(), ...) — records
// logged without a span in their context pass through unchanged.
//
// # Relationship to go-cqrs-lite's otel module
//
// Both modules expose a Setup with the same shape (WithService,
// WithSpanExporter, WithoutGlobalRegistration, ...); a process uses exactly
// ONE of them — both register the process-global providers. This module
// exists so plain HTTP services need no go-cqrs-lite dependency; when you
// already use go-cqrs-lite, call its Setup (or this one — the resource,
// propagator, and provider mechanics are equivalent) and wire BOTH
// instrumentations against the same globals.
package otel
