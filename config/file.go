package config

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	goFasterYaml "github.com/go-faster/yaml"
	errorfamily "github.com/larsartmann/go-error-family"
)

// readFile loads a YAML or JSON file into the flat key space the file layer
// uses: lowercased, dotted struct paths ("nested.addr"), with scalar values
// rendered as the same raw strings the env parser consumes. Nested maps are
// flattened; lists become comma-joined strings.
func readFile(path string) (map[string]string, []error) {
	ext := strings.ToLower(filepath.Ext(path))

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, []error{&FieldError{Field: fileToken, Key: path, Kind: KindFileRead, Err: err}}
	}

	var raw map[string]any

	var decodeErr error

	switch ext {
	case ".yaml", ".yml":
		decodeErr = goFasterYaml.Unmarshal(data, &raw)
	case ".json":
		decodeErr = json.Unmarshal(data, &raw)
	default:
		return nil, []error{&FieldError{
			Field: fileToken, Key: path, Kind: KindUnknownExtension,
			Err: errorfamily.NewRejection("config.file_extension", "file layer supports .yaml, .yml, and .json"),
		}}
	}

	if decodeErr != nil {
		return nil, []error{&FieldError{Field: fileToken, Key: path, Kind: KindFileDecode, Err: decodeErr}}
	}

	return flatten(raw, "")
}

// flatten walks a decoded file tree into dotted lowercase keys.
func flatten(raw map[string]any, prefix string) (map[string]string, []error) {
	out := make(map[string]string, len(raw))

	var errs []error

	for key, val := range raw {
		full := strings.ToLower(key)
		if prefix != "" {
			full = prefix + "." + strings.ToLower(key)
		}

		switch typed := val.(type) {
		case map[string]any:
			nested, nestedErrs := flatten(typed, full)
			errs = append(errs, nestedErrs...)

			maps.Copy(out, nested)
		default:
			rendered, err := valueToString(typed)
			if err != nil {
				errs = append(errs, &FieldError{Field: fileToken, Key: full, Kind: KindFileDecode, Err: err})

				continue
			}

			out[full] = rendered
		}
	}

	return out, errs
}

// valueToString renders a decoded scalar (or list of scalars) as the raw
// string the shared parser consumes. Maps are handled by flatten; nil and
// anything exotic are rejected — the file layer never guesses types.
func valueToString(val any) (string, error) {
	switch typed := val.(type) {
	case string:
		return typed, nil
	case bool:
		return strconv.FormatBool(typed), nil
	case int:
		return strconv.Itoa(typed), nil
	case int64:
		return strconv.FormatInt(typed, 10), nil
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32), nil
	case []any:
		parts := make([]string, 0, len(typed))

		for _, item := range typed {
			rendered, err := valueToString(item)
			if err != nil {
				return "", err
			}

			parts = append(parts, rendered)
		}

		return strings.Join(parts, ","), nil
	case nil:
		return "", errorfamily.NewRejection("config.null_file_value", "null value: set a concrete value or remove the key")
	default:
		return "", errorfamily.Newf(errorfamily.Rejection, "config.file_value_type", "unsupported file value type %T", val)
	}
}
