package appkit

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
)

// MetricsConfig enables the opt-in Prometheus surface on a [Service]:
//
//	cfg.Metrics = &appkit.MetricsConfig{
//	    Path:           "/metrics",
//	    BasicAuthUser:  "metrics",
//	    BasicAuthPass:  os.Getenv("METRICS_PASS"),
//	}
//
// The exposition is pure Prometheus text format with ZERO dependencies —
// no prometheus client, no otel. The exported metric names are a stable
// contract (dashboards and alerts compile against them):
//
//	appkit_http_request_duration_seconds  histogram {method, route, status}
//	appkit_http_responses_total           counter   {method, route, status}
//	appkit_http_requests_in_flight        gauge
//	appkit_build_info                     gauge=1   {version}
//
// # The `_ratio` exporter trap
//
// OTEL's Prometheus exporter appends `_ratio` to metrics whose unit is `1`.
// A dashboard built on this surface's plain names will NOT match names
// produced by the otel module's exporter for unit-1 metrics (and vice
// versa). When migrating between the two surfaces, diff the metric names —
// do not assume a 1:1 rename.
//
// # Authentication is mandatory by default
//
// Unauthenticated metrics endpoints leak request cardinality, route
// topology, and build metadata — and Gatus/Prometheus scrapes are exactly
// how an attacker discovers them. Either set BasicAuthUser/BasicAuthPass,
// or set AllowUnauthenticated explicitly (loopback-only deployments, or a
// proxy that enforces auth in front).
type MetricsConfig struct {
	// Path is the scrape endpoint. Default: "/metrics".
	Path string

	// BasicAuthUser and BasicAuthPass enable HTTP Basic Auth on the
	// endpoint. Comparison is constant-time. When both are empty,
	// AllowUnauthenticated must be explicitly true or NewService rejects
	// the configuration.
	BasicAuthUser string
	BasicAuthPass string

	// AllowUnauthenticated serves the endpoint without auth. Opt-in only:
	// use it when a proxy in front enforces authentication, or for
	// loopback-only services.
	AllowUnauthenticated bool
}

// applyMetricsDefaults fills zero values.
func (c *MetricsConfig) applyMetricsDefaults() {
	if c.Path == "" {
		c.Path = "/metrics"
	}
}

// validate rejects unsafe metric configurations at construction time.
func (c *MetricsConfig) validate() error {
	if !strings.HasPrefix(c.Path, "/") {
		return fmt.Errorf("%w: metrics path %q must start with '/'", errInvalidMetricsConfig, c.Path)
	}

	if c.BasicAuthUser == "" && c.BasicAuthPass == "" && !c.AllowUnauthenticated {
		return fmt.Errorf(
			"%w: metrics endpoint %s is unauthenticated — set BasicAuthUser/BasicAuthPass, "+
				"or set AllowUnauthenticated explicitly (a proxy enforces auth, or loopback-only deployment)",
			errInvalidMetricsConfig, c.Path)
	}

	return nil
}

// errInvalidMetricsConfig rejects unsafe metrics configurations.
var errInvalidMetricsConfig = errorfamily.NewRejection(
	"appkit.metrics_config_invalid", "invalid metrics configuration")

// durationBuckets are the request-duration histogram buckets in seconds
// (semconv-style 0..10s coverage).
//
//nolint:gochecknoglobals // constant bucket set, shared across collectors
var durationBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

// metricsCollector aggregates request metrics and renders Prometheus text.
// Every exported metric name it emits is part of the stable contract
// documented on [MetricsConfig].
type metricsCollector struct {
	mu        sync.Mutex
	byRoute   map[string]*routeSeries // key: method|route|status
	inFlight  int64
	version   string
	startTime time.Time
}

// routeSeries holds the histogram + counter for one
// (method, route, status) series. Duration samples are bucketed, not
// stored — bounded memory, exact Prometheus histogram semantics.
type routeSeries struct {
	buckets []uint64 // len(durationBuckets)+1 (last is +Inf)
	count   uint64
	sum     float64
}

func newMetricsCollector(version string) *metricsCollector {
	return &metricsCollector{ //nolint:exhaustruct_v5 // mu zero value is ready; inFlight starts at 0
		byRoute:   make(map[string]*routeSeries),
		startTime: time.Now(),
		version:   version,
	}
}

// middleware returns the collection middleware. Route labels come from the
// ServeMux pattern (r.Pattern is stamped by the mux on the request it is
// handed, so it is readable after next returns); unmatched paths fall back
// to "unmatched" so cardinality stays bounded.
func (m *metricsCollector) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		m.inFlight++
		m.mu.Unlock()

		start := time.Now()
		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
			wroteHeader:    false,
		}
		next.ServeHTTP(rec, r)

		duration := time.Since(start).Seconds()

		route := r.Pattern
		if route == "" {
			route = "unmatched"
		}

		m.mu.Lock()
		defer m.mu.Unlock()

		m.inFlight--

		key := r.Method + "|" + route + "|" + strconv.Itoa(rec.status)

		series, ok := m.byRoute[key]
		if !ok {
			series = &routeSeries{
				buckets: make([]uint64, len(durationBuckets)+1),
				count:   0,
				sum:     0,
			}
			m.byRoute[key] = series
		}

		series.count++
		series.sum += duration

		for i, bound := range durationBuckets {
			if duration <= bound {
				series.buckets[i]++

				break
			}
		}

		series.buckets[len(durationBuckets)]++
	})
}

// statusRecorder captures the response status without buffering the body.
type statusRecorder struct {
	http.ResponseWriter

	status      int
	wroteHeader bool
}

func (s *statusRecorder) WriteHeader(code int) {
	if !s.wroteHeader {
		s.status = code
		s.wroteHeader = true
	}

	s.ResponseWriter.WriteHeader(code)
}

// WriteHeader implements http.ResponseWriter; the flusher passthrough keeps
// SSE streams working under the metrics middleware.
func (s *statusRecorder) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// handler serves the Prometheus text exposition.
func (m *metricsCollector) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		var b strings.Builder

		m.mu.Lock()
		m.writeBuildInfo(&b)
		m.writeInFlight(&b)
		m.writeHistogram(&b)
		m.writeResponseTotals(&b)
		m.mu.Unlock()

		_, _ = w.Write([]byte(b.String()))
	})
}

func (m *metricsCollector) writeBuildInfo(b *strings.Builder) {
	b.WriteString("# HELP appkit_build_info Build information. The value is always 1.\n")
	b.WriteString("# TYPE appkit_build_info gauge\n")
	fmt.Fprintf(b, "appkit_build_info{version=%q} 1\n", escapeLabelValue(m.version))
}

func (m *metricsCollector) writeInFlight(b *strings.Builder) {
	b.WriteString("# HELP appkit_http_requests_in_flight Requests currently being served.\n")
	b.WriteString("# TYPE appkit_http_requests_in_flight gauge\n")
	fmt.Fprintf(b, "appkit_http_requests_in_flight %d\n", m.inFlight)
}

// seriesKeyParts is the number of fields in a byRoute key (method|route|status).
const seriesKeyParts = 3

func (m *metricsCollector) writeHistogram(b *strings.Builder) {
	b.WriteString("# HELP appkit_http_request_duration_seconds HTTP request duration by route.\n")
	b.WriteString("# TYPE appkit_http_request_duration_seconds histogram\n")

	for _, key := range m.sortedRoutes() {
		parts := strings.SplitN(key, "|", seriesKeyParts)
		series := m.byRoute[key]

		labels := `method="` + escapeLabelValue(parts[0]) + `",route="` + escapeLabelValue(parts[1]) + `"`
		statusLabel := `,status="` + parts[2] + `"`

		var cumulative uint64
		for i, bound := range durationBuckets {
			cumulative = series.buckets[i]
			fmt.Fprintf(b, "appkit_http_request_duration_seconds_bucket{%s%s,le=\"%s\"} %d\n",
				labels, statusLabel, formatBound(bound), cumulative)
		}

		fmt.Fprintf(b, "appkit_http_request_duration_seconds_bucket{%s%s,le=\"+Inf\"} %d\n",
			labels, statusLabel, series.buckets[len(durationBuckets)])
		fmt.Fprintf(b, "appkit_http_request_duration_seconds_sum{%s%s} %g\n", labels, statusLabel, series.sum)
		fmt.Fprintf(b, "appkit_http_request_duration_seconds_count{%s%s} %d\n", labels, statusLabel, series.count)
	}
}

func (m *metricsCollector) writeResponseTotals(b *strings.Builder) {
	b.WriteString("# HELP appkit_http_responses_total HTTP responses by route and status.\n")
	b.WriteString("# TYPE appkit_http_responses_total counter\n")

	for _, key := range m.sortedRoutes() {
		parts := strings.SplitN(key, "|", seriesKeyParts)

		fmt.Fprintf(b, "appkit_http_responses_total{method=%q,route=%q,status=%q} %d\n",
			escapeLabelValue(parts[0]), escapeLabelValue(parts[1]), parts[2], m.byRoute[key].count)
	}
}

// sortedRoutes fixes the exposition order so scrapes are diffable.
func (m *metricsCollector) sortedRoutes() []string {
	keys := make([]string, 0, len(m.byRoute))
	for key := range m.byRoute {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func formatBound(b float64) string {
	return strconv.FormatFloat(b, 'g', -1, 64)
}

// escapeLabelValue escapes the characters Prometheus forbids inside quoted
// label values.
func escapeLabelValue(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `"`, `\"`)
	v = strings.ReplaceAll(v, "\n", `\n`)

	return v
}

// metricsAuth wraps the exposition handler with the configured Basic Auth.
// Both sides are hashed before comparison so the comparison is
// constant-time regardless of input length.
func metricsAuth(cfg *MetricsConfig, next http.Handler) http.Handler {
	if cfg.BasicAuthUser == "" && cfg.BasicAuthPass == "" {
		return next
	}

	expectedHash := sha256.Sum256([]byte(cfg.BasicAuthUser + ":" + cfg.BasicAuthPass))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		presentedHash := sha256.Sum256([]byte(user + ":" + pass))

		if !ok || subtle.ConstantTimeCompare(presentedHash[:], expectedHash[:]) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="metrics"`)
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		next.ServeHTTP(w, r)
	})
}
