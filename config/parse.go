package config

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
)

var durationType = reflect.TypeFor[time.Duration]() //nolint:gochecknoglobals // reflect type, immutable by construction

// supportedType reports whether parseValue can bind fieldType. The set is
// frozen for v0.1.0: string, bool, all int/uint widths, float32/64,
// time.Duration, and []string. Anything else tagged `env` fails Load with
// KindUnsupportedType — no silent skips, no WeaklyTypedInput-style coercion.
func supportedType(fieldType reflect.Type) bool {
	if fieldType == durationType {
		return true
	}

	switch fieldType.Kind() { //nolint:exhaustive // scalar kinds enumerated; struct handled by caller
	case reflect.String,
		reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.Slice:
		return fieldType.Elem().Kind() == reflect.String
	default:
		return false
	}
}

// parseValue parses raw into a settable value of type fieldType. Every layer
// parses from its own raw string, so a default, a file value, an env value,
// and a flag value all flow through the exact same parser — one place to get
// boundary semantics right.
func parseValue(fieldType reflect.Type, raw string) (reflect.Value, error) {
	if fieldType == durationType {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return reflect.Value{}, parseErr(raw, err)
		}

		return reflect.ValueOf(parsed).Convert(fieldType), nil
	}

	if fieldType.Kind() == reflect.Slice {
		parts := strings.Split(raw, ",")

		for index, part := range parts {
			parts[index] = strings.TrimSpace(part)
		}

		return reflect.ValueOf(parts).Convert(fieldType), nil
	}

	return parseScalar(fieldType, raw)
}

// parseScalar handles the non-duration, non-slice kinds.
func parseScalar(fieldType reflect.Type, raw string) (reflect.Value, error) {
	switch fieldType.Kind() { //nolint:exhaustive // scalar kinds enumerated; default is the error
	case reflect.String:
		return reflect.ValueOf(raw).Convert(fieldType), nil
	case reflect.Bool:
		return parseAndConvert(fieldType, raw, strconv.ParseBool)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(raw, parseBase, fieldType.Bits())
		if err != nil {
			return reflect.Value{}, parseErr(raw, err)
		}

		return reflect.ValueOf(parsed).Convert(fieldType), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(raw, parseBase, fieldType.Bits())
		if err != nil {
			return reflect.Value{}, parseErr(raw, err)
		}

		return reflect.ValueOf(parsed).Convert(fieldType), nil
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(raw, fieldType.Bits())
		if err != nil {
			return reflect.Value{}, parseErr(raw, err)
		}

		return reflect.ValueOf(parsed).Convert(fieldType), nil
	default:
		return reflect.Value{}, errorfamily.Newf(
			errorfamily.Rejection, "config.unsupported_field_type",
			"unsupported field type %s", fieldType,
		)
	}
}

// parseBase is the fixed numeric base for int/uint fields (decimal only —
// octal-looking env values like "0755" must fail, not silently mean 493).
const parseBase = 10

func parseAndConvert(
	fieldType reflect.Type,
	raw string,
	parse func(string) (bool, error),
) (reflect.Value, error) {
	parsed, err := parse(raw)
	if err != nil {
		return reflect.Value{}, parseErr(raw, err)
	}

	return reflect.ValueOf(parsed).Convert(fieldType), nil
}

// parseErr adds the raw value to the parser's message so the FieldError line
// is actionable without re-running the load.
func parseErr(raw string, err error) error {
	return errorfamily.WrapRejectionf(err, "config.parse", "cannot parse %q", raw)
}
