package config

import (
	"reflect"
	"strings"
	"time"
)

// fieldSpec is one env-tagged leaf field of the target struct, with the
// reflection index and every derived key the layers address it by.
type fieldSpec struct {
	// path is the struct path ("Nested.Addr"); for promoted fields from
	// embedded structs it is the bare leaf name ("Addr"), matching how Go
	// promotes embedded fields.
	path string
	// fileKey is the lowercased path with dots ("nested.addr") — the file
	// layer's key space.
	fileKey string
	// envKey is the full environment key ("APP_ADDR"); never empty for a
	// spec — untagged fields produce no spec at all.
	envKey string
	// defaultVal is the raw `default` tag value.
	defaultVal string
	hasDefault bool
	required   bool
	index      []int
	typ        reflect.Type
}

// nestedWalk reports whether the field is a struct the collector must walk
// into, and with which prefixes. Two shapes recurse: an untagged embedded
// struct (its exported leaves promote as top-level fields, prefixes
// unchanged) and a tagged nested struct (its child tags carry relative
// suffixes under the tag prefix). Everything else — untagged named structs,
// untagged scalars, tagged leaves — is not walked.
func nestedWalk(
	field reflect.StructField,
	envTag string,
	hasTag, isStruct bool,
	envPrefix, pathPrefix string,
) (string, string, bool) {
	switch {
	case field.Anonymous && !hasTag:
		return envPrefix, pathPrefix, true
	case isStruct && hasTag:
		name, _, _ := strings.Cut(envTag, ",")

		return envPrefix + name + "_", joinPath(pathPrefix, field.Name), true
	default:
		return "", "", false
	}
}

// collectFields walks t and produces one spec per env-tagged leaf field.
//
// Rules (frozen in the design note):
//   - fields without an `env` tag are never touched by any layer;
//   - `env:"PREFIX"` on a struct field makes child keys PREFIX_CHILD;
//   - an embedded (anonymous) struct field without its own tag is walked with
//     the enclosing prefix — its leaves bind as promoted top-level fields;
//   - a leaf with an env tag must be a supported scalar type, else a
//     KindUnsupportedType error is returned.
func collectFields(
	t reflect.Type,
	envPrefix string,
	pathPrefix string,
	index []int,
) ([]fieldSpec, []error) {
	var specs []fieldSpec

	var errs []error

	for i := range t.NumField() {
		field := t.Field(i)

		if field.PkgPath != "" && !field.Anonymous {
			continue // unexported field; an unexported EMBEDDED type still promotes its exported fields
		}

		envTag, hasTag := field.Tag.Lookup("env")
		name, opts, _ := strings.Cut(envTag, ",")

		idx := append(append([]int{}, index...), i)

		isStruct := field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeFor[time.Time]()

		childEnvPrefix, childPathPrefix, walk := nestedWalk(field, envTag, hasTag, isStruct, envPrefix, pathPrefix)

		if walk {
			childSpecs, childErrs := collectFields(field.Type, childEnvPrefix, childPathPrefix, idx)
			specs, errs = append(specs, childSpecs...), append(errs, childErrs...)

			continue
		}

		if isStruct || !hasTag {
			continue // untagged named struct or untagged scalar: the caller's own wiring — never touch
		}

		if !supportedType(field.Type) {
			errs = append(
				errs,
				&FieldError{ //nolint:exhaustruct_v5 // Err deliberately nil: the kind carries the failure
					Field: joinPath(pathPrefix, field.Name), Key: name,
					Kind: KindUnsupportedType,
				},
			)

			continue
		}

		spec := fieldSpec{ //nolint:exhaustruct_v5 // fileKey/default filled just below
			path:     joinPath(pathPrefix, field.Name),
			envKey:   envPrefix + name,
			index:    idx,
			typ:      field.Type,
			required: hasOption(opts, "required"),
		}

		if def, ok := field.Tag.Lookup("default"); ok {
			spec.defaultVal, spec.hasDefault = def, true
		}

		spec.fileKey = strings.ToLower(spec.path)

		specs = append(specs, spec)
	}

	return specs, errs
}

// joinPath builds the struct path; the first segment has no leading dot.
func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}

	return prefix + "." + name
}

func hasOption(opts, want string) bool {
	for opt := range strings.SplitSeq(opts, ",") {
		if strings.TrimSpace(opt) == want {
			return true
		}
	}

	return false
}
