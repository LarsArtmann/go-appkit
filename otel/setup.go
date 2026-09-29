package otel

import (
	"context"
	"errors"
	"fmt"
	"io"

	errorfamily "github.com/larsartmann/go-error-family"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// SetupOption configures the provider setup.
type SetupOption func(*setupConfig)

type setupConfig struct {
	serviceName            string
	serviceVersion         string
	instanceID             string
	environment            string
	spanExporter           sdktrace.SpanExporter
	sampler                sdktrace.Sampler
	metricReader           sdkmetric.Reader
	propagator             propagation.TextMapPropagator
	stdoutWriter           io.Writer
	otlp                   *otlpConfig
	skipGlobalRegistration bool
}

// WithService identifies the service in telemetry via resource attributes.
// serviceName is required for meaningful traces; version and instanceID are
// optional (pass "" to omit). When WithService is not used, OTEL_SERVICE_NAME
// supplies the name without code changes.
func WithService(name, version, instanceID string) SetupOption {
	return func(c *setupConfig) {
		c.serviceName = name
		c.serviceVersion = version
		c.instanceID = instanceID
	}
}

// WithEnvironment records the deployment environment ("production",
// "staging", ...) as the deployment.environment resource attribute — the
// dimension SigNoz filters services and exceptions by.
func WithEnvironment(name string) SetupOption {
	return func(c *setupConfig) {
		c.environment = name
	}
}

// WithSpanExporter attaches a span exporter (OTLP, stdout, etc.).
// Without one, spans are recorded but not exported — unless the standard
// OTEL_EXPORTER_OTLP_* environment variables ask for OTLP, in which case
// Setup builds the OTLP exporters itself (see [WithOTLP]).
func WithSpanExporter(e sdktrace.SpanExporter) SetupOption {
	return func(c *setupConfig) {
		c.spanExporter = e
	}
}

// WithSampler overrides the default sampler (ParentBased AlwaysSample).
// Typical production choices: sdktrace.TraceIDRatioBased(0.1) for head
// sampling, or a tail-based sampler making the decision per-span.
// When omitted, the sampler itself is env-native: OTEL_TRACES_SAMPLER and
// OTEL_TRACES_SAMPLER_ARG configure sampling without code changes (e.g.
// parentbased_traceidratio + 0.1), and invalid values fall back to the
// default via the OTel global error handler.
func WithSampler(s sdktrace.Sampler) SetupOption {
	return func(c *setupConfig) {
		c.sampler = s
	}
}

// WithMetricReader attaches a metric reader (OTLP, prometheus, stdout,
// manual, etc.). When omitted, no metric reader is configured and metrics
// instruments become no-ops.
func WithMetricReader(r sdkmetric.Reader) SetupOption {
	return func(c *setupConfig) {
		c.metricReader = r
	}
}

// WithPropagator overrides the default W3C (trace-context + baggage)
// propagator.
func WithPropagator(p propagation.TextMapPropagator) SetupOption {
	return func(c *setupConfig) {
		c.propagator = p
	}
}

// WithStdoutExporter pretty-prints spans to the given writer. Ideal for
// local development — pass os.Stdout to see traces in your terminal. The
// exporter is constructed internally; for custom stdout configuration use
// WithSpanExporter.
func WithStdoutExporter(w io.Writer) SetupOption {
	return func(c *setupConfig) {
		c.stdoutWriter = w
	}
}

// WithoutGlobalRegistration skips registering the providers as the
// process-wide global TracerProvider, MeterProvider, and TextMapPropagator.
// Use this when you need an isolated Provider — e.g. in tests, or when
// running multiple services in one process where each owns its providers.
// The returned Provider is fully functional; only the otel.Set* globals are
// skipped, so otel.GetTracerProvider() is left unchanged.
func WithoutGlobalRegistration() SetupOption {
	return func(c *setupConfig) {
		c.skipGlobalRegistration = true
	}
}

// Provider wraps the TracerProvider and MeterProvider with a unified
// Shutdown. Its Shutdown method matches the signature expected by
// appkit.ServiceConfig.ShutdownHooks:
//
//	cfg.ShutdownHooks = []func(context.Context) error{provider.Shutdown}
type Provider struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
}

// AsTracerProvider returns the underlying OTel TracerProvider.
func (p *Provider) AsTracerProvider() *sdktrace.TracerProvider {
	return p.tracerProvider
}

// AsMeterProvider returns the underlying OTel MeterProvider.
func (p *Provider) AsMeterProvider() *sdkmetric.MeterProvider {
	return p.meterProvider
}

// Classified sentinels: shutdown and SDK-build failures are environmental
// (Infrastructure), so consumers get HTTPStatus 503 and correct retry
// semantics instead of an unclassified error. Messages are pinned — tests
// and grep-based ops runbooks match on them.
var (
	errShutdown    = errorfamily.NewInfrastructure("otel.shutdown_incomplete", "otel provider shutdown incomplete")
	errBuildRes    = errorfamily.NewInfrastructure("otel.resource_build_failed", "failed to build OTel resource")
	errStdoutSetup = errorfamily.NewInfrastructure("otel.stdout_exporter_failed", "failed to build stdout exporter")
)

// Shutdown flushes pending spans and metrics, then releases resources.
// Always call this on application exit — via ServiceConfig.ShutdownHooks in
// an appkit service, which runs it after the server released its
// connections.
//
// A ForceFlush precedes each provider's Shutdown: spans finished moments
// before shutdown can still sit in the batch processor's asynchronous
// queue, and Shutdown alone does not guarantee they reach the exporter.
func (p *Provider) Shutdown(ctx context.Context) error {
	var errs []error

	err := p.tracerProvider.ForceFlush(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("tracer flush: %w", err))
	}

	err = p.tracerProvider.Shutdown(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("tracer shutdown: %w", err))
	}

	err = p.meterProvider.ForceFlush(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("meter flush: %w", err))
	}

	err = p.meterProvider.Shutdown(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("meter shutdown: %w", err))
	}

	if len(errs) > 0 {
		return errors.Join(append([]error{errShutdown}, errs...)...)
	}

	return nil
}

// Setup creates and registers a TracerProvider and MeterProvider in one
// call. It configures the W3C propagator (trace-context + baggage),
// HTTP-optimized histogram views, and a resource identifying the service.
// The returned Provider owns both providers.
//
// Typical usage:
//
//	provider, err := appkitotel.Setup(
//	    appkitotel.WithService("orders-api", "1.0.0", "instance-1"),
//	    appkitotel.WithOTLP(),
//	)
//	if err != nil {
//	    return err
//	}
//	cfg := appkit.DefaultServiceConfig()
//	cfg.ShutdownHooks = []func(context.Context) error{provider.Shutdown}
//
// # Environment-driven setup (zero code)
//
// Export is off until something asks for it. Asking happens in code
// ([WithOTLP], [WithSpanExporter], [WithStdoutExporter], [WithMetricReader])
// or through the environment: when OTEL_EXPORTER_OTLP_ENDPOINT (or a
// signal-specific OTEL_EXPORTER_OTLP_TRACES_ENDPOINT /
// OTEL_EXPORTER_OTLP_METRICS_ENDPOINT) is set, Setup builds the OTLP/HTTP
// exporters for the affected signals itself. Explicit code options win per
// signal. Combined with OTEL_SERVICE_NAME and OTEL_TRACES_SAMPLER, a bare
// Setup() call is fully deployment-configurable.
//
// The global TracerProvider, MeterProvider, and propagator are set so
// [Middleware] picks them up automatically; pass WithoutGlobalRegistration
// to keep the process globals untouched.
func Setup(opts ...SetupOption) (*Provider, error) {
	cfg := &setupConfig{} //nolint:exhaustruct_v5 // options applied below

	for _, opt := range opts {
		opt(cfg)
	}

	if err := applyOTLPWiring(context.Background(), cfg); err != nil {
		return nil, err
	}

	res, err := buildResource(cfg)
	if err != nil {
		return nil, err
	}

	spanExporter, err := resolveSpanExporter(cfg)
	if err != nil {
		return nil, err
	}

	tracerProvider := buildTracerProvider(cfg, res, spanExporter)
	meterProvider := buildMeterProvider(cfg, res)

	if !cfg.skipGlobalRegistration {
		otel.SetTextMapPropagator(propagatorOrDefault(cfg.propagator))
		otel.SetTracerProvider(tracerProvider)
		otel.SetMeterProvider(meterProvider)
	}

	return &Provider{tracerProvider: tracerProvider, meterProvider: meterProvider}, nil
}

// resolveSpanExporter picks the configured span exporter, falling back to
// the stdout exporter when one was requested. Without either, spans are
// recorded but not exported.
func resolveSpanExporter(cfg *setupConfig) (sdktrace.SpanExporter, error) {
	if cfg.spanExporter != nil {
		return cfg.spanExporter, nil
	}

	if cfg.stdoutWriter == nil {
		return nil, nil
	}

	exporter, err := stdouttrace.New(
		stdouttrace.WithWriter(cfg.stdoutWriter),
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errStdoutSetup, err)
	}

	return exporter, nil
}

// buildTracerProvider assembles the tracer provider. An explicit sampler
// overrides the env-derived one (the SDK reads OTEL_TRACES_SAMPLER itself
// when WithSampler is not passed); with neither, the SDK default
// ParentBased(AlwaysSample) applies.
func buildTracerProvider(
	cfg *setupConfig,
	res *resource.Resource,
	exporter sdktrace.SpanExporter,
) *sdktrace.TracerProvider {
	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
	}

	if cfg.sampler != nil {
		opts = append(opts, sdktrace.WithSampler(cfg.sampler))
	}

	if exporter != nil {
		opts = append(opts, sdktrace.WithBatcher(exporter))
	}

	return sdktrace.NewTracerProvider(opts...)
}

// buildMeterProvider assembles the meter provider with the HTTP views; the
// reader may be nil (instruments become no-ops).
func buildMeterProvider(cfg *setupConfig, res *resource.Resource) *sdkmetric.MeterProvider {
	opts := []sdkmetric.Option{
		sdkmetric.WithResource(res),
		sdkmetric.WithView(NewHTTPViews()...),
	}

	if cfg.metricReader != nil {
		opts = append(opts, sdkmetric.WithReader(cfg.metricReader))
	}

	return sdkmetric.NewMeterProvider(opts...)
}

// propagatorOrDefault resolves the propagation set: the configured one when
// present, else the W3C default.
func propagatorOrDefault(p propagation.TextMapPropagator) propagation.TextMapPropagator {
	if p != nil {
		return p
	}

	return NewTextMapPropagator()
}

// applyOTLPWiring fills the unset exporter slots from OTLP: an explicit
// [WithOTLP] always wires both signals (unless the signal has its own
// explicit option); otherwise the environment decides per signal. This is
// the zero-code path — the deployment env turns telemetry on and off.
// Construction failures surface to Setup's caller: a misconfigured exporter
// must fail the process at startup, not silently drop telemetry.
func applyOTLPWiring(ctx context.Context, cfg *setupConfig) error {
	explicit := cfg.otlp != nil

	if cfg.spanExporter == nil && cfg.stdoutWriter == nil && (explicit || otlpTracesEnvConfigured()) {
		exp, err := newOTLPSpanExporter(ctx, cfg.otlp)
		if err != nil {
			return err
		}

		cfg.spanExporter = exp
	}

	if cfg.metricReader == nil && (explicit || otlpMetricsEnvConfigured()) {
		reader, err := newOTLPMetricReader(ctx, cfg.otlp)
		if err != nil {
			return err
		}

		cfg.metricReader = reader
	}

	return nil
}

// buildResource assembles the OTel resource: the SDK standard detectors
// (environment variables incl. OTEL_SERVICE_NAME, telemetry SDK, host) fill
// the base, and the code-configured service identity wins on conflict.
func buildResource(cfg *setupConfig) (*resource.Resource, error) {
	var attrs []attribute.KeyValue

	if cfg.serviceName != "" {
		attrs = ServiceResourceAttributes(cfg.serviceName, cfg.serviceVersion, cfg.instanceID)
	}

	if cfg.environment != "" {
		attrs = append(attrs, semconv.DeploymentEnvironment(cfg.environment))
	}

	res, err := resource.New(
		context.Background(),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(attrs...),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errBuildRes, err)
	}

	return res, nil
}
