package otel

import (
	"context"
	"errors"
	"fmt"
	"io"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// SetupOption configures the provider setup.
type SetupOption func(*setupConfig)

type setupConfig struct {
	serviceName            string
	serviceVersion         string
	instanceID             string
	spanExporter           sdktrace.SpanExporter
	sampler                sdktrace.Sampler
	metricReader           sdkmetric.Reader
	propagator             propagation.TextMapPropagator
	stdoutWriter           io.Writer
	skipGlobalRegistration bool
}

// WithService identifies the service in telemetry via resource attributes.
// serviceName is required for meaningful traces; version and instanceID are
// optional (pass "" to omit).
func WithService(name, version, instanceID string) SetupOption {
	return func(c *setupConfig) {
		c.serviceName = name
		c.serviceVersion = version
		c.instanceID = instanceID
	}
}

// WithSpanExporter attaches a span exporter (OTLP, stdout, etc.).
// Without one, spans are recorded but not exported — useful for
// in-memory testing. See WithStdoutExporter for the common development
// case.
func WithSpanExporter(e sdktrace.SpanExporter) SetupOption {
	return func(c *setupConfig) {
		c.spanExporter = e
	}
}

// WithSampler overrides the default sampler (ParentBased AlwaysSample).
// Typical production choices: sdktrace.TraceIDRatioBased(0.1) for head
// sampling, or a tail-based sampler making the decision per-span.
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

var (
	errShutdown    = errors.New("otel provider shutdown incomplete")
	errBuildRes    = errors.New("failed to build OTel resource")
	errStdoutSetup = errors.New("failed to build stdout exporter")
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
//	    appkitotel.WithSpanExporter(otlpExporter),
//	)
//	if err != nil {
//	    return err
//	}
//	cfg := appkit.DefaultServiceConfig()
//	cfg.ShutdownHooks = []func(context.Context) error{provider.Shutdown}
//
// Without a span exporter, spans are recorded but not exported — ideal for
// in-memory testing. The global TracerProvider, MeterProvider, and
// propagator are set so [Middleware] picks them up automatically; pass
// WithoutGlobalRegistration to keep the process globals untouched.
func Setup(opts ...SetupOption) (*Provider, error) {
	cfg := &setupConfig{} //nolint:exhaustruct_v5 // options applied below

	for _, opt := range opts {
		opt(cfg)
	}

	res, err := buildResource(cfg)
	if err != nil {
		return nil, err
	}

	spanExporter := cfg.spanExporter
	if spanExporter == nil && cfg.stdoutWriter != nil {
		spanExporter, err = stdouttrace.New(
			stdouttrace.WithWriter(cfg.stdoutWriter),
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errStdoutSetup, err)
		}
	}

	propagator := cfg.propagator
	if propagator == nil {
		propagator = NewTextMapPropagator()
	}

	if !cfg.skipGlobalRegistration {
		otel.SetTextMapPropagator(propagator)
	}

	sampler := sdktrace.ParentBased(sdktrace.AlwaysSample())
	if cfg.sampler != nil {
		sampler = cfg.sampler
	}

	tpOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	}

	if spanExporter != nil {
		tpOpts = append(tpOpts, sdktrace.WithBatcher(spanExporter))
	}

	tracerProvider := sdktrace.NewTracerProvider(tpOpts...)

	mpOpts := []sdkmetric.Option{
		sdkmetric.WithResource(res),
		sdkmetric.WithView(NewHTTPViews()...),
	}

	if cfg.metricReader != nil {
		mpOpts = append(mpOpts, sdkmetric.WithReader(cfg.metricReader))
	}

	meterProvider := sdkmetric.NewMeterProvider(mpOpts...)

	if !cfg.skipGlobalRegistration {
		otel.SetTracerProvider(tracerProvider)
		otel.SetMeterProvider(meterProvider)
	}

	return &Provider{tracerProvider: tracerProvider, meterProvider: meterProvider}, nil
}

// buildResource assembles the OTel resource from the configured service
// identity plus the SDK's standard detectors (environment, telemetry SDK).
func buildResource(cfg *setupConfig) (*resource.Resource, error) {
	attrs := ServiceResourceAttributes(cfg.serviceName, cfg.serviceVersion, cfg.instanceID)

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(attrs...),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errBuildRes, err)
	}

	return res, nil
}
