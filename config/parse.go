package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var durationType = reflect.TypeFor[time.Duration]() //nolint:gochecknoglobals // reflect type, immutable by construction

// supportedType reports whether parseValue can bind ty. The set is frozen for
// v0.1.0: string, bool, all int/uint widths, float32/64, time.Duration, and
// []string. Anything else tagged `env` fails Load with KindUnsupportedType —
// no silent skips, no WeaklyTypedInput-style coercion.
func supportedType(ty reflect.Type) bool {
	if ty == durationType {
		return true
	}

	switch ty.Kind() { //nolint:exhaustive // scalar kinds enumerated; struct handled by caller
	case reflect.String,
		reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.Slice:
		return ty.Elem().Kind() == reflect.String
	default:
		return false
	}
}

// parseValue parses raw into a settable value of type ty. Every layer parses
// from its own raw string, so a default, a file value, an env value, and a
// flag value all flow through the exact same parser — one place to get
// boundary semantics right.
func parseValue(ty reflect.Type, raw string) (reflect.Value, error) {
	if ty == durationType {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return reflect.Value{}, parseErr(raw, err)
		}

		return reflect.ValueOf(d).Convert(ty), nil
	}

	if ty.Kind() == reflect.Slice {
		parts := strings.Split(raw, ",")

		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}

		return reflect.ValueOf(parts).Convert(ty), nil
	}

	switch ty.Kind() { //nolint:exhaustive // scalar kinds enumerated; default is the error
	case reflect.String:
		return reflect.ValueOf(raw).Convert(ty), nil
	case reflect.Bool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return reflect.Value{}, parseErr(raw, err)
		}

		return reflect.ValueOf(v).Convert(ty), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(raw, 10, ty.Bits())
		if err != nil {
			return reflect.Value{}, parseErr(raw, err)
		}

		return reflect.ValueOf(v).Convert(ty), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(raw, 10, ty.Bits())
		if err != nil {
			return reflect.Value{}, parseErr(raw, err)
		}

		return reflect.ValueOf(v).Convert(ty), nil
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(raw, ty.Bits())
		if err != nil {
			return reflect.Value{}, parseErr(raw, err)
		}

		return reflect.ValueOf(v).Convert(ty), nil
	default:
		return reflect.Value{}, fmt.Errorf("unsupported field type %s", ty)
	}
}

// parseErr adds the raw value to the parser's message so the FieldError line
// is actionable without re-running the load.
func parseErr(raw string, err error) error {
	return fmt.Errorf("cannot parse %q: %w", raw, err)
}
