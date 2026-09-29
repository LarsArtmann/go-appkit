package otel

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Standard OTLP environment variables (OpenTelemetry specification) that
// gate the env-driven setup path. When OTEL_EXPORTER_OTLP_ENDPOINT (or a
// signal-specific override) is set, Setup exports via OTLP without any code
// change — the deployment environment turns telemetry on and off.
const (
	envOTLPEndpoint        = "OTEL_EXPORTER_OTLP_ENDPOINT"
	envOTLPTracesEndpoint  = "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"
	envOTLPMetricsEndpoint = "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"
)

// OTLPOption configures the OTLP/HTTP exporters built by [WithOTLP].
type OTLPOption func(*otlpConfig)

// otlpConfig carries the explicit OTLP wiring. Unset fields defer to the
// standard OTEL_EXPORTER_OTLP_* environment variables, which the exporters
// read natively (endpoint, headers, timeout, compression, retry policy).
type otlpConfig struct {
	endpoint string
	headers  map[string]string
	timeout  time.Duration
}

// WithOTLPEndpoint sets the collector base URL, including the scheme —
// "http://localhost:4318" for a local SigNoz, "https://ingest.<region>.signoz.cloud"
// for SigNoz Cloud. An "http://" scheme sends insecure OTLP/HTTP; "https://"
// upgrades to TLS. When omitted, the exporters resolve the endpoint from
// OTEL_EXPORTER_OTLP_ENDPOINT (or the signal-specific variables).
func WithOTLPEndpoint(url string) OTLPOption {
	return func(c *otlpConfig) {
		c.endpoint = url
	}
}

// WithOTLPHeaders sets headers sent with every OTLP request — how SigNoz
// Cloud authenticates ingestion ({"signoz-ingestion-key": "<key>"}).
// OTEL_EXPORTER_OTLP_HEADERS supplies further headers without code changes.
func WithOTLPHeaders(headers map[string]string) OTLPOption {
	return func(c *otlpConfig) {
		c.headers = headers
	}
}

// WithOTLPTimeout bounds each OTLP export request (default: the exporter's
// own default of 10s; OTEL_EXPORTER_OTLP_TIMEOUT overrides without code).
func WithOTLPTimeout(d time.Duration) OTLPOption {
	return func(c *otlpConfig) {
		c.timeout = d
	}
}

// WithOTLP wires OTLP/HTTP export for both signals — the production path to
// SigNoz (or any OTLP backend) in one option:
//
//	provider, err := appkitotel.Setup(
//	    appkitotel.WithService("orders-api", "1.0.0", pod),
//	    appkitotel.WithOTLP(), // endpoint from OTEL_EXPORTER_OTLP_ENDPOINT
//	)
//
// Traces export via a batch span processor, metrics via a periodic reader
// (interval: OTEL_METRIC_EXPORT_INTERVAL, default 60s). Explicit
// [WithSpanExporter] / [WithMetricReader] options take precedence for their
// signal, so WithOTLP composes — e.g. OTLP traces alongside a Prometheus
// metric reader. Without [WithOTLPEndpoint], the standard OTEL_* environment
// variables configure the exporters natively.
func WithOTLP(opts ...OTLPOption) SetupOption {
	return func(setup *setupConfig) {
		cfg := &otlpConfig{} //nolint:exhaustruct_v5 // options applied below
		for _, opt := range opts {
			opt(cfg)
		}

		setup.otlp = cfg
	}
}

// otlpTracesEnvConfigured reports whether the environment asks for OTLP
// trace export: the signal shares OTEL_EXPORTER_OTLP_ENDPOINT unless
// OTEL_EXPORTER_OTLP_TRACES_ENDPOINT overrides it.
func otlpTracesEnvConfigured() bool {
	return envSet(envOTLPEndpoint) || envSet(envOTLPTracesEndpoint)
}

// otlpMetricsEnvConfigured is the metric-signal counterpart of
// [otlpTracesEnvConfigured].
func otlpMetricsEnvConfigured() bool {
	return envSet(envOTLPEndpoint) || envSet(envOTLPMetricsEndpoint)
}

func envSet(key string) bool {
	return os.Getenv(key) != ""
}

// newOTLPSpanExporter builds the OTLP/HTTP span exporter from the explicit
// config; unset fields fall through to the exporters' native env handling.
func newOTLPSpanExporter(ctx context.Context, cfg *otlpConfig) (sdktrace.SpanExporter, error) {
	opts := otlpExporterOptions(cfg, func(raw string) otlptracehttp.Option {
		return otlptracehttp.WithEndpointURL(otlpSignalURL(raw, "/v1/traces"))
	}, otlptracehttp.WithHeaders, otlptracehttp.WithTimeout)

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errOTLPSetup, err)
	}

	return exporter, nil
}

// newOTLPMetricReader builds the periodic OTLP/HTTP metric reader; the
// export interval itself is env-native (OTEL_METRIC_EXPORT_INTERVAL).
func newOTLPMetricReader(ctx context.Context, cfg *otlpConfig) (sdkmetric.Reader, error) {
	opts := otlpExporterOptions(cfg, func(raw string) otlpmetrichttp.Option {
		return otlpmetrichttp.WithEndpointURL(otlpSignalURL(raw, "/v1/metrics"))
	}, otlpmetrichttp.WithHeaders, otlpmetrichttp.WithTimeout)

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errOTLPSetup, err)
	}

	return sdkmetric.NewPeriodicReader(exporter), nil
}

// otlpSignalURL joins the configured base endpoint with the standard signal
// path ("http://collector:4318" + "/v1/traces"). A path on the endpoint URL
// becomes a base prefix ("https://gw/corp" + "/v1/traces") — the same
// base-URL-plus-signal-path convention SigNoz's own docs use. Unparseable
// URLs pass through unchanged so the exporter surfaces the problem.
func otlpSignalURL(raw, signalPath string) string {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return raw
	}

	parsed.Path = strings.TrimSuffix(parsed.Path, "/") + signalPath

	return parsed.String()
}

// otlpExporterOptions maps the shared config into either exporter's option
// type. The generic parameter collapsed three near-identical option lists
// into one; the endpoint mapping stays per-exporter because the signal
// paths differ (traces vs metrics).
func otlpExporterOptions[O any](
	cfg *otlpConfig,
	endpoint func(string) O,
	headers func(map[string]string) O,
	timeout func(time.Duration) O,
) []O {
	var opts []O

	if cfg == nil {
		return nil
	}

	if cfg.endpoint != "" {
		opts = append(opts, endpoint(cfg.endpoint))
	}

	if len(cfg.headers) > 0 {
		opts = append(opts, headers(cfg.headers))
	}

	if cfg.timeout > 0 {
		opts = append(opts, timeout(cfg.timeout))
	}

	return opts
}

// errOTLPSetup classifies exporter construction failures: reaching the
// configured collector is an environmental concern, so consumers get the
// Infrastructure family (HTTP 503) and standard retry semantics.
var errOTLPSetup = errorfamily.NewInfrastructure("otel.otlp_exporter_failed", "failed to build OTLP exporter")
