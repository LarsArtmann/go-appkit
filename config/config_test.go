package config

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
)

type basicConfig struct {
	Addr     string        `default:":8080" env:"TEST_ADDR"`
	LogLevel string        `default:"info"  env:"TEST_LOG_LEVEL"`
	APIKey   string        `                env:"TEST_API_KEY,required"`
	TTL      time.Duration `default:"60s"   env:"TEST_TTL"`
	Retries  int           `default:"3"     env:"TEST_RETRIES"`
	Verbose  bool          `                env:"TEST_VERBOSE"`
	Hosts    []string      `                env:"TEST_HOSTS"`
}

type scalarsConfig struct {
	Text    string        `env:"SC_TEXT"`
	Flag    bool          `env:"SC_FLAG"`
	Small   int8          `env:"SC_SMALL"`
	Counter uint16        `env:"SC_COUNTER"`
	Ratio   float64       `env:"SC_RATIO"`
	Span    time.Duration `env:"SC_SPAN"`
}

type nestedConfig struct {
	Sub struct {
		Addr string `env:"ADDR"`
	} `env:"NESTED"`
	Top string `env:"NESTED_TOP"`
}

type embeddedBase struct {
	BaseAddr string `default:"embedded-default" env:"EMB_BASE_ADDR"`
}

type embeddedConfig struct {
	embeddedBase

	Own string `env:"EMB_OWN"`
}

type unsupportedConfig struct {
	Bad map[string]string `env:"UNSUPPORTED_BAD"`
}

func envMap(vals map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := vals[key]

		return v, ok
	}
}

func kindsOf(err error) map[string]FieldErrorKind {
	kinds := map[string]FieldErrorKind{}
	for _, fe := range FieldErrorsOf(err) {
		kinds[fe.Field] = fe.Kind
	}

	return kinds
}

func TestLoad_DefaultsApplied(t *testing.T) {
	t.Parallel()

	cfg, err := Load[basicConfig](WithLookup(envMap(map[string]string{"TEST_API_KEY": "k"})))
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

func TestLoad_ScalarTypesRoundTrip(t *testing.T) {
	t.Parallel()

	cfg, err := Load[scalarsConfig](WithLookup(envMap(map[string]string{
		"SC_TEXT":    "hello",
		"SC_FLAG":    "true",
		"SC_SMALL":   "-12",
		"SC_COUNTER": "4000",
		"SC_RATIO":   "3.14",
		"SC_SPAN":    "1h30m",
	})))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Text != "hello" || !cfg.Flag || cfg.Small != -8-4 || cfg.Counter != 4000 || cfg.Ratio != 3.14 {
		t.Errorf("scalar round trip wrong: %+v", cfg)
	}

	if cfg.Span != 90*time.Minute {
		t.Errorf("Span = %v, want 1h30m", cfg.Span)
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
	if !slices.Equal(cfg.Hosts, want) {
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

	if rejection.Family() != errorfamily.Rejection {
		t.Errorf("family = %v, want Rejection", rejection.Family())
	}

	fieldErrs := FieldErrorsOf(err)
	if len(fieldErrs) != 1 {
		t.Fatalf("FieldErrorsOf = %v, want exactly one", fieldErrs)
	}

	fe := fieldErrs[0]
	if fe.Field != "APIKey" || fe.Key != "TEST_API_KEY" {
		t.Errorf("FieldError = %s/%s, want APIKey/TEST_API_KEY", fe.Field, fe.Key)
	}

	if fe.Kind != KindMissingRequired {
		t.Errorf("Kind = %s, want missing_required", fe.Kind)
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

	kinds := kindsOf(err)
	if kinds["TTL"] != KindParseFailed {
		t.Errorf("TTL kind = %v, want parse_failed (all: %v)", kinds["TTL"], kinds)
	}

	if kinds["Retries"] != KindParseFailed {
		t.Errorf("Retries kind = %v, want parse_failed", kinds["Retries"])
	}
}

func TestLoad_BadDefaultCaughtAtLoad(t *testing.T) {
	t.Parallel()

	_, err := Load[struct {
		TTL time.Duration `default:"bogus" env:"BADDEF_TTL"`
	}](WithLookup(envMap(map[string]string{})))
	if err == nil {
		t.Fatal("bad default must fail Load")
	}

	if kindsOf(err)["TTL"] != KindParseFailed {
		t.Errorf("kinds = %v, want parse_failed for TTL", kindsOf(err))
	}
}

func TestLoad_PrecedenceMatrix(t *testing.T) {
	t.Parallel()

	type target struct {
		Addr string `default:"default-value" env:"PREC_ADDR"`
	}

	writeFile := func(t *testing.T, content string) string {
		t.Helper()

		path := filepath.Join(t.TempDir(), "app.yaml")
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}

		return path
	}

	tests := []struct {
		name  string
		file  string
		env   map[string]string
		flags map[string]string
		want  string
	}{
		{
			name: "only default applies",
			env:  map[string]string{},
			want: "default-value",
		},
		{
			name: "file beats default",
			file: "addr: file-value",
			env:  map[string]string{},
			want: "file-value",
		},
		{
			name: "env beats file",
			file: "addr: file-value",
			env:  map[string]string{"PREC_ADDR": "env-value"},
			want: "env-value",
		},
		{
			name:  "flag beats env",
			file:  "addr: file-value",
			env:   map[string]string{"PREC_ADDR": "env-value"},
			flags: map[string]string{"Addr": "flag-value"},
			want:  "flag-value",
		},
		{
			name:  "empty flag falls through to env",
			file:  "addr: file-value",
			env:   map[string]string{"PREC_ADDR": "env-value"},
			flags: map[string]string{"Addr": ""},
			want:  "env-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var fileOpt Option

			if tt.file != "" {
				fileOpt = WithFile(writeFile(t, tt.file))
			} else {
				noop := func(*loaderConfig) {}
				fileOpt = noop
			}

			cfg, err := Load[target](fileOpt, WithLookup(envMap(tt.env)), WithValues(tt.flags))
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if cfg.Addr != tt.want {
				t.Errorf("Addr = %q, want %q", cfg.Addr, tt.want)
			}
		})
	}
}

func TestLoad_FileYAML(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "app.yaml")
	content := `
addr: ":7000"
ttl: 90s
retries: 9
hosts:
  - a.example.com
  - b.example.com
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load[basicConfig](
		WithFile(path),
		WithLookup(envMap(map[string]string{"TEST_API_KEY": "k"})),
	)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Addr != ":7000" || cfg.TTL != 90*time.Second || cfg.Retries != 9 {
		t.Errorf("file values not applied: %+v", cfg)
	}

	want := []string{"a.example.com", "b.example.com"}
	if !slices.Equal(cfg.Hosts, want) {
		t.Errorf("Hosts = %v, want %v", cfg.Hosts, want)
	}
}

func TestLoad_FileJSONEnvOverridesFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "app.json")
	if err := os.WriteFile(path, []byte(`{"addr": ":7000", "ttl": "90s"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load[basicConfig](
		WithFile(path),
		WithLookup(envMap(map[string]string{
			"TEST_API_KEY": "k",
			"TEST_ADDR":    ":8000",
		})),
	)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Addr != ":8000" {
		t.Errorf("Addr = %q, want env value :8000", cfg.Addr)
	}

	if cfg.TTL != 90*time.Second {
		t.Errorf("TTL = %v, want file value 90s", cfg.TTL)
	}
}

func TestLoad_FileNestedYAML(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "nested.yaml")
	if err := os.WriteFile(path, []byte("sub:\n  addr: nested-value\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load[nestedConfig](
		WithFile(path),
		WithLookup(envMap(map[string]string{})),
	)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Sub.Addr != "nested-value" {
		t.Errorf("Sub.Addr = %q, want nested-value", cfg.Sub.Addr)
	}
}

func TestLoad_FileUnknownKeyIsStrict(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "unknown.yaml")
	if err := os.WriteFile(path, []byte("typo_key: value\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load[basicConfig](WithFile(path), WithLookup(envMap(map[string]string{"TEST_API_KEY": "k"})))
	if err == nil {
		t.Fatal("unknown file key must fail Load")
	}

	found := false

	for _, fe := range FieldErrorsOf(err) {
		if fe.Kind == KindFileDecode && fe.Key == "typo_key" {
			found = true
		}
	}

	if !found {
		t.Errorf("no strict unknown-key error: %v", FieldErrorsOf(err))
	}
}

func TestLoad_FileUnknownExtension(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "app.toml")
	if err := os.WriteFile(path, []byte("addr = x"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load[basicConfig](WithFile(path), WithLookup(envMap(map[string]string{"TEST_API_KEY": "k"})))
	if kindsOf(err)["(file)"] != KindUnknownExtension {
		t.Errorf("kinds = %v, want unknown_extension for (file)", kindsOf(err))
	}
}

func TestLoad_FileMissing(t *testing.T) {
	t.Parallel()

	_, err := Load[basicConfig](
		WithFile(filepath.Join(t.TempDir(), "absent.yaml")),
		WithLookup(envMap(map[string]string{"TEST_API_KEY": "k"})),
	)
	if kindsOf(err)["(file)"] != KindFileRead {
		t.Errorf("kinds = %v, want file_read for (file)", kindsOf(err))
	}
}

func TestLoad_NestedEnvPrefixes(t *testing.T) {
	t.Parallel()

	cfg, err := Load[nestedConfig](WithLookup(envMap(map[string]string{
		"NESTED_ADDR": ":6000",
		"NESTED_TOP":  "top",
	})))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Sub.Addr != ":6000" || cfg.Top != "top" {
		t.Errorf("nested prefixes wrong: %+v", cfg)
	}
}

func TestLoad_EmbeddedFieldsBindPromoted(t *testing.T) {
	t.Parallel()

	cfg, err := Load[embeddedConfig](WithLookup(envMap(map[string]string{
		"EMB_BASE_ADDR": ":5000",
		"EMB_OWN":       "own",
	})))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseAddr != ":5000" {
		t.Errorf("BaseAddr = %q, want :5000 (promoted env binding)", cfg.BaseAddr)
	}

	if cfg.Own != "own" {
		t.Errorf("Own = %q, want own", cfg.Own)
	}

	cfgDefaults, err := Load[embeddedConfig](WithLookup(envMap(map[string]string{"EMB_OWN": "own"})))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfgDefaults.BaseAddr != "embedded-default" {
		t.Errorf("BaseAddr = %q, want embedded default", cfgDefaults.BaseAddr)
	}
}

func TestLoad_EmbeddedFileKeyIsPromoted(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "embedded.yaml")
	if err := os.WriteFile(path, []byte("baseaddr: :4000\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load[embeddedConfig](WithFile(path), WithLookup(envMap(map[string]string{"EMB_OWN": "own"})))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseAddr != ":4000" {
		t.Errorf("BaseAddr = %q, want promoted file key :4000", cfg.BaseAddr)
	}
}

func TestLoad_UnsupportedTaggedType(t *testing.T) {
	t.Parallel()

	_, err := Load[unsupportedConfig](WithLookup(envMap(map[string]string{"UNSUPPORTED_BAD": "x"})))
	if kindsOf(err)["Bad"] != KindUnsupportedType {
		t.Errorf("kinds = %v, want unsupported_type for Bad", kindsOf(err))
	}
}

func TestLoad_WithValuesFlagLayer(t *testing.T) {
	t.Parallel()

	cfg, err := Load[basicConfig](
		WithLookup(envMap(map[string]string{
			"TEST_API_KEY": "k",
			"TEST_ADDR":    "env-value",
		})),
		WithValues(map[string]string{"addr": "flag-value"}),
	)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Addr != "flag-value" {
		t.Errorf("Addr = %q, want flag-value (case-insensitive field match)", cfg.Addr)
	}
}

func TestLoad_WithValuesUnknownField(t *testing.T) {
	t.Parallel()

	_, err := Load[basicConfig](
		WithLookup(envMap(map[string]string{"TEST_API_KEY": "k"})),
		WithValues(map[string]string{"Nope": "x"}),
	)
	if kindsOf(err)["Nope"] != KindUnknownField {
		t.Errorf("kinds = %v, want unknown_field for Nope", kindsOf(err))
	}
}

func TestLoad_ValidateHookRunsLast(t *testing.T) {
	t.Parallel()

	_, err := Load[basicConfig](
		WithLookup(envMap(map[string]string{"TEST_API_KEY": "k"})),
		WithValidate(func(c *basicConfig) error {
			if c.TTL <= time.Minute {
				return errors.New("ttl at or below one minute")
			}

			return nil
		}),
	)
	if kindsOf(err)["(validate)"] != KindValidation {
		t.Errorf("kinds = %v, want validation for (validate)", kindsOf(err))
	}

	cfg, err := Load[basicConfig](
		WithLookup(envMap(map[string]string{"TEST_API_KEY": "k", "TEST_TTL": "2m"})),
		WithValidate(func(c *basicConfig) error { return nil }),
	)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.TTL != 2*time.Minute {
		t.Errorf("TTL = %v, want 2m", cfg.TTL)
	}
}

func TestLoad_NonStructTypeParameterFails(t *testing.T) {
	t.Parallel()

	_, err := Load[int]()
	if err == nil {
		t.Fatal("Load[int]() must fail")
	}
}

func TestKeys_ListsEveryEnvKey(t *testing.T) {
	t.Parallel()

	keys, err := Keys[basicConfig]()
	if err != nil {
		t.Fatalf("Keys() error = %v", err)
	}

	want := []string{
		"TEST_ADDR", "TEST_API_KEY", "TEST_HOSTS", "TEST_LOG_LEVEL",
		"TEST_RETRIES", "TEST_TTL", "TEST_VERBOSE",
	}
	if !slices.Equal(keys, want) {
		t.Errorf("Keys() = %v, want %v", keys, want)
	}
}

func TestKeys_NestedPrefixes(t *testing.T) {
	t.Parallel()

	keys, err := Keys[nestedConfig]()
	if err != nil {
		t.Fatalf("Keys() error = %v", err)
	}

	want := []string{"NESTED_ADDR", "NESTED_TOP"}
	if !slices.Equal(keys, want) {
		t.Errorf("Keys() = %v, want %v", keys, want)
	}
}

func TestLoad_DefaultLookupUsesProcessEnv(t *testing.T) {
	t.Setenv("TEST_API_KEY", "from-process-env")
	t.Setenv("TEST_ADDR", ":3000")

	cfg, err := Load[basicConfig]()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.APIKey != "from-process-env" || cfg.Addr != ":3000" {
		t.Errorf("process env not consulted: %+v", cfg)
	}
}

func TestLoader_Reusable(t *testing.T) {
	t.Parallel()

	loader := NewLoader[basicConfig](WithLookup(envMap(map[string]string{"TEST_API_KEY": "k"})))

	first, err := loader.Load()
	if err != nil {
		t.Fatalf("first Load() error = %v", err)
	}

	second, err := loader.Load()
	if err != nil {
		t.Fatalf("second Load() error = %v", err)
	}

	if first == second {
		t.Error("each Load must produce a fresh value")
	}
}
