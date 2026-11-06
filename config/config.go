package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"

	errorfamily "github.com/larsartmann/go-error-family"
)

// FieldErrorKind classifies why a configuration field failed to load.
type FieldErrorKind string

const (
	// KindMissingRequired: a field tagged `env:"...,required"` resolved to its
	// zero value after every layer ran.
	KindMissingRequired FieldErrorKind = "missing_required"

	// KindParseFailed: a raw value (env, file, flag, or default) could not be
	// parsed into the field's type.
	KindParseFailed FieldErrorKind = "parse_failed"

	// KindUnsupportedType: a field carries an `env` tag but its type is not
	// one of the supported scalars.
	KindUnsupportedType FieldErrorKind = "unsupported_type"

	// KindFileRead: the configured file could not be read.
	KindFileRead FieldErrorKind = "file_read"

	// KindFileDecode: the file could not be decoded, holds a value no field
	// can consume, or contains a key no tagged field declares.
	KindFileDecode FieldErrorKind = "file_decode"

	// KindUnknownExtension: the file path's extension is neither .yaml/.yml
	// nor .json.
	KindUnknownExtension FieldErrorKind = "unknown_extension"

	// KindUnknownField: a WithValues entry names a field the type does not
	// declare.
	KindUnknownField FieldErrorKind = "unknown_field"

	// KindValidation: the consumer's WithValidate hook rejected the loaded
	// config.
	KindValidation FieldErrorKind = "validation"
)

// FieldError is the typed, per-field failure. Load returns a single
// go-error-family Rejection wrapping one FieldError per problem, so consumers
// can either read the joined message or errors.As their way to programmatic
// field paths.
type FieldError struct {
	// Field is the struct path of the offending field, e.g. "Nested.Addr".
	Field string
	// Key is the environment key, file key, or flag name that failed.
	Key string
	// Kind classifies the failure.
	Kind FieldErrorKind
	// Err is the underlying cause, if any.
	Err error
}

// Error renders the field, its key, and the reason on one line.
func (e *FieldError) Error() string {
	reason := string(e.Kind)
	if e.Err != nil {
		reason = e.Err.Error()
	}

	if e.Key == "" {
		return fmt.Sprintf("%s: %s", e.Field, reason)
	}

	return fmt.Sprintf("%s (%s): %s", e.Field, e.Key, reason)
}

// Unwrap exposes the underlying cause for errors.Is/errors.As.
func (e *FieldError) Unwrap() error { return e.Err }

// Load resolves a T from defaults, an optional file, the environment, and
// optional flag results, in that order of increasing precedence (see the
// design note: flag > env > file > default). It returns nil and a single
// Rejection (code "config.invalid") wrapping one FieldError per problem when
// anything fails; a returned error always means the config is unusable.
func Load[T any](opts ...Option) (*T, error) {
	return NewLoader[T](opts...).Load()
}

// Loader is a reusable Load: options are fixed at construction, Load may be
// called repeatedly (the SIGHUP-reload shape; a validate-gated hot-reload
// wrapper is the demand-gated F4 battery, not this type).
type Loader[T any] struct {
	cfg loaderConfig
}

type loaderConfig struct {
	file     string
	lookup   func(string) (string, bool)
	values   map[string]string
	validate func(any) error
}

// NewLoader builds a Loader from the given options.
func NewLoader[T any](opts ...Option) *Loader[T] {
	cfg := loaderConfig{lookup: os.LookupEnv}

	for _, opt := range opts {
		opt(&cfg)
	}

	return &Loader[T]{cfg: cfg}
}

// Option configures a Loader.
type Option func(*loaderConfig)

// WithFile adds the file layer: path must end in .yaml, .yml, or .json; its
// keys are lowercased field names (nested sections use dotted paths, e.g.
// "nested.addr"). File values sit between defaults and the environment in
// precedence. Optional — the layer is off without it.
func WithFile(path string) Option {
	return func(c *loaderConfig) { c.file = path }
}

// WithLookup replaces the environment source (default os.LookupEnv). Values of
// exactly "" count as unset, matching the fleet's envOr convention. Use it in
// tests instead of mutating the process environment.
func WithLookup(fn func(string) (string, bool)) Option {
	return func(c *loaderConfig) { c.lookup = fn }
}

// WithValues is the flag layer: field name (case-insensitive; dotted path for
// nested fields) to raw string value, applied with the highest precedence.
// Empty values count as unset, so cmdguard-style flags with "" defaults pass
// through cleanly when not provided.
func WithValues(vals map[string]string) Option {
	return func(c *loaderConfig) {
		if c.values == nil {
			c.values = make(map[string]string, len(vals))
		}

		for k, v := range vals {
			c.values[k] = v
		}
	}
}

// WithValidate runs fn against the fully resolved config, after every layer
// and the required checks. Its error is wrapped as a KindValidation FieldError
// inside the standard Rejection. Use it for cross-field rules; single-field
// rules belong in the field's parser or the required tag.
func WithValidate[T any](fn func(*T) error) Option {
	return func(c *loaderConfig) {
		c.validate = func(v any) error { return fn(v.(*T)) }
	}
}

// Load resolves the config per the Loader's options. See the package-level
// [Load] for the precedence and error contract.
func (l *Loader[T]) Load() (*T, error) {
	out := new(T)

	rv := reflect.ValueOf(out).Elem()

	if rv.Kind() != reflect.Struct {
		return nil, errorfamily.Newf(
			errorfamily.Rejection,
			"config.invalid",
			"type parameter must be a struct, got %s", rv.Kind(),
		)
	}

	specs, errs := collectFields(rv.Type(), "", "", nil)
	if len(errs) > 0 {
		return nil, rejectionFrom(errs)
	}

	l.applyDefaults(rv, specs, &errs)
	l.applyFile(rv, specs, &errs)
	l.applyEnv(rv, specs, &errs)
	l.applyValues(rv, specs, &errs)
	checkRequired(rv, specs, &errs)
	l.runValidate(out, &errs)

	if len(errs) > 0 {
		return nil, rejectionFrom(errs)
	}

	return out, nil
}

func (l *Loader[T]) applyDefaults(rv reflect.Value, specs []fieldSpec, errs *[]error) {
	for _, spec := range specs {
		if !spec.hasDefault {
			continue
		}

		setParsed(rv, spec, spec.defaultVal, spec.envKey, errs)
	}
}

func (l *Loader[T]) applyFile(rv reflect.Value, specs []fieldSpec, errs *[]error) {
	if l.cfg.file == "" {
		return
	}

	vals, fileErrs := readFile(l.cfg.file)
	*errs = append(*errs, fileErrs...)
	if len(fileErrs) > 0 {
		return
	}

	byFileKey := make(map[string]fieldSpec, len(specs))
	for _, spec := range specs {
		byFileKey[spec.fileKey] = spec
	}

	for key, raw := range vals {
		spec, ok := byFileKey[key]
		if !ok {
			*errs = append(*errs, &FieldError{
				Field: "(file)", Key: key, Kind: KindFileDecode,
				Err: errors.New("no tagged field declares this key"),
			})

			continue
		}

		setParsed(rv, spec, raw, key, errs)
	}
}

func (l *Loader[T]) applyEnv(rv reflect.Value, specs []fieldSpec, errs *[]error) {
	for _, spec := range specs {
		if spec.envKey == "" || l.cfg.lookup == nil {
			continue
		}

		raw, ok := l.cfg.lookup(spec.envKey)
		if !ok || raw == "" {
			continue
		}

		setParsed(rv, spec, raw, spec.envKey, errs)
	}
}

func (l *Loader[T]) applyValues(rv reflect.Value, specs []fieldSpec, errs *[]error) {
	for name, raw := range l.cfg.values {
		if raw == "" {
			continue
		}

		spec, ok := specByPath(specs, name)
		if !ok {
			*errs = append(*errs, &FieldError{
				Field: name, Key: name, Kind: KindUnknownField,
				Err: errors.New("no tagged field with this name"),
			})

			continue
		}

		setParsed(rv, spec, raw, name, errs)
	}
}

func (l *Loader[T]) runValidate(out *T, errs *[]error) {
	if l.cfg.validate == nil {
		return
	}

	if err := l.cfg.validate(out); err != nil {
		*errs = append(*errs, &FieldError{
			Field: "(validate)", Key: "validate", Kind: KindValidation, Err: err,
		})
	}
}

// checkRequired flags tagged-required fields that still hold their zero value.
func checkRequired(rv reflect.Value, specs []fieldSpec, errs *[]error) {
	for _, spec := range specs {
		if !spec.required {
			continue
		}

		if rv.FieldByIndex(spec.index).IsZero() {
			*errs = append(*errs, &FieldError{
				Field: spec.path, Key: spec.envKey, Kind: KindMissingRequired,
				Err: errors.New("required value is unset or empty"),
			})
		}
	}
}

func specByPath(specs []fieldSpec, name string) (fieldSpec, bool) {
	for _, spec := range specs {
		if strings.EqualFold(spec.path, name) {
			return spec, true
		}
	}

	return fieldSpec{}, false
}

// setParsed parses raw into spec's field and sets it, appending a KindParseFailed
// FieldError on failure.
func setParsed(rv reflect.Value, spec fieldSpec, raw string, key string, errs *[]error) {
	v, err := parseValue(spec.typ, raw)
	if err != nil {
		*errs = append(*errs, &FieldError{Field: spec.path, Key: key, Kind: KindParseFailed, Err: err})

		return
	}

	rv.FieldByIndex(spec.index).Set(v)
}

// rejectionFrom folds FieldErrors (and any other error) into the single
// Rejection every Load failure returns.
func rejectionFrom(errs []error) error {
	return errorfamily.WrapRejection(errors.Join(errs...), "config.invalid", errors.Join(errs...).Error())
}

// Keys returns every environment key T declares via `env` tags, sorted. It is
// the push-model answer to koanf's hidden-registry problem: the key set is a
// fact of the type, computable without loading. Table tests can iterate it to
// prove every key is actually wired (set junk, expect a parse failure).
func Keys[T any]() ([]string, error) {
	specs, errs := collectFields(reflect.TypeOf(*new(T)), "", "", nil)
	if len(errs) > 0 {
		return nil, rejectionFrom(errs)
	}

	keys := make([]string, 0, len(specs))
	for _, spec := range specs {
		if spec.envKey != "" {
			keys = append(keys, spec.envKey)
		}
	}

	slices.Sort(keys)

	return keys, nil
}

// FieldErrorsOf collects every FieldError inside an error returned by Load
// (or Keys), walking the whole wrap/join tree. Nil for errors from other
// sources.
func FieldErrorsOf(err error) []*FieldError {
	var out []*FieldError

	var walk func(error)

	walk = func(e error) {
		if e == nil {
			return
		}

		var fieldErr *FieldError

		if errors.As(e, &fieldErr) {
			out = append(out, fieldErr)

			return
		}

		switch unwrapper := e.(type) {
		case interface{ Unwrap() error }:
			walk(unwrapper.Unwrap())
		case interface{ Unwrap() []error }:
			for _, child := range unwrapper.Unwrap() {
				walk(child)
			}
		}
	}

	walk(err)

	return out
}
