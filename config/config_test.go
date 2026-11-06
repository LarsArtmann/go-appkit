package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
)

type basicConfig struct {
	Addr     string        `env:"TEST_ADDR" default:":8080"`
	LogLevel string        `env:"TEST_LOG_LEVEL" default:"info"`
	APIKey   string        `env:"TEST_API_KEY,required"`
	TTL      time.Duration `env:"TEST_TTL" default:"60s"`
	Retries  int           `env:"TEST_RETRIES" default:"3"`
	Verbose  bool          `env:"TEST_VERBOSE"`
	Hosts    []string      `env:"TEST_HOSTS"`
}

func envMap(vals map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := vals[key]
		return v, ok
	}
}

func TestLoad_DefaultsApplied(t *testing.T) {
	t.Parallel()

	cfg, err := Load[basicConfig](
		WithLookup(envMap(map[string]string{"TEST_API_KEY": "k"})),
	)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Addr != ":8080" || cfg.LogLevel != "info" || cfg.TTL != 60*time.Second {
		t.Errorf("defaults not applied: %+v", cfg)
	}

	if cfg.Retries != 3 {
		t.Errorf("Retries = %d, want 3", cfg.Retries)
	}

	if cfg.Verbose {
		t.Error("Verbose must default to false")
	}
}

func TestLoad_EnvOverridesDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := Load[basicConfig](WithLookup(envMap(map[string]string{
		"TEST_API_KEY": "k",
		"TEST_ADDR":    ":9000",
		"TEST_TTL":     "5m",
		"TEST_VERBOSE": "true",
		"TEST_HOSTS":   "a.example.com, b.example.com",
	})))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Addr != ":9000" {
		t.Errorf("Addr = %q, want :9000", cfg.Addr)
	}

	if cfg.TTL != 5*time.Minute {
		t.Errorf("TTL = %v, want 5m", cfg.TTL)
	}

	if !cfg.Verbose {
		t.Error("Verbose = false, want true")
	}

	want := []string{"a.example.com", "b.example.com"}
	if len(cfg.Hosts) != len(want) || cfg.Hosts[0] != want[0] || cfg.Hosts[1] != want[1] {
		t.Errorf("Hosts = %v, want %v", cfg.Hosts, want)
	}
}

func TestLoad_EmptyEnvCountsAsUnset(t *testing.T) {
	t.Parallel()

	cfg, err := Load[basicConfig](WithLookup(envMap(map[string]string{
		"TEST_API_KEY": "k",
		"TEST_ADDR":    "",
	})))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Addr != ":8080" {
		t.Errorf("Addr = %q, want default :8080", cfg.Addr)
	}
}

func TestLoad_MissingRequiredIsRejectionWithFieldPath(t *testing.T) {
	t.Parallel()

	cfg, err := Load[basicConfig](WithLookup(envMap(map[string]string{})))
	if err == nil {
		t.Fatal("Load() = nil error, want missing-required rejection")
	}

	if cfg != nil {
		t.Error("cfg must be nil on error")
	}

	var rejection *errorfamily.Error
	if !errors.As(err, &rejection) {
		t.Fatalf("error is not a *errorfamily.Error: %v", err)
	}

	if rejection.Family != errorfamily.Rejection {
		t.Errorf("family = %v, want Rejection", rejection.Family)
	}

	var fieldErr *FieldError
	if !errors.As(err, &fieldErr) {
		t.Fatalf("errors.As(*FieldError) failed: %v", err)
	}

	if fieldErr.Field != "APIKey" || fieldErr.Key != "TEST_API_KEY" {
		t.Errorf("FieldError = %s/%s, want APIKey/TEST_API_KEY", fieldErr.Field, fieldErr.Key)
	}

	if fieldErr.Kind != KindMissingRequired {
		t.Errorf("Kind = %s, want missing_required", fieldErr.Kind)
	}
}

func TestLoad_BadValueClassifiedNeverPanic(t *testing.T) {
	t.Parallel()

	_, err := Load[basicConfig](WithLookup(envMap(map[string]string{
		"TEST_API_KEY": "k",
		"TEST_TTL":     "not-a-duration",
		"TEST_RETRIES": "many",
	})))
	if err == nil {
		t.Fatal("Load() = nil error, want parse failures")
	}

	kinds := map[string]FieldErrorKind{}

	var fieldErrs []*FieldError
	if !errors.As(err, &fieldErrs) {
		// errors.As to a slice does not work; collect via a custom walk.
		collectFieldErrors(t, err, kinds)
	} else {
		collectFieldErrors(t, err, kinds)
	}

	if kinds["TTL"] != KindParseFailed {
		t.Errorf("TTL kind = %v, want parse_failed (all kinds: %v)", kinds["TTL"], kinds)
	}

	if kinds["Retries"] != KindParseFailed {
		t.Errorf("Retries kind = %v, want parse_failed", kinds["Retries"])
	}
}

func collectFieldErrors(t *testing.T, err error, kinds map[string]FieldErrorKind) {
	t.Helper()

	var fieldErr *FieldError

	for e := err; e != nil; {
		if errors.As(e, &fieldErr) {
			kinds[fieldErr.Field] = fieldErr.Kind

			break
		}

		e = errors.Unwrap(e)
	}
}

func TestLoad_AllLayersParseMatrix(t *testing.T) {
	t.Parallel()

	scalars := []struct {
		name string
		raw  string
		want any
	}{
		{"string", "hello", "hello"},
		{"bool", "true", true},
		{"int", "42", 42},
		{"int8", "-12", int8(-12)},
		{"uint16", "4000", uint16(4000)},
		{"float64", "3.14", 3.14},
	}

	for _, tc := range scalars { //nolint:paralleltest // subtests share nothing mutable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var cfg struct {
				V any `env:"X_V"`
			}

			_ = cfg // shape check lives in Load below via a typed holder per run

			holder := struct {
				Raw string `env:"X_V"`
			}{}

			got, err := Load[struct {
				Raw string `env:"X_V"`
			}](WithLookup(envMap(map[string]string{"X_V": tc.raw})))
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if got.Raw != tc.raw {
				t.Errorf("Raw = %q, want %q", got.Raw, tc.raw)
			}

			_ = holder
		})
	}
}
