package security

import (
	"net/url"
	"strings"
	"unicode"

	"github.com/microcosm-cc/bluemonday"
)

// strictPolicy strips all HTML elements and attributes, leaving only plain
// text. Compiled once; bluemonday policies are safe for concurrent use.
//
//nolint:gochecknoglobals // compiled policy, safe to share across goroutines
var strictPolicy = bluemonday.StrictPolicy()

// SanitizeText strips HTML tags, control characters (preserving tab, CR,
// LF), and leading/trailing whitespace from a user-supplied text field. Use
// it on every string field that enters the system from an HTTP request body
// before storing or re-displaying it.
func SanitizeText(input string) string {
	trimmed := strings.TrimSpace(input)
	sanitized := strictPolicy.Sanitize(trimmed)

	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r' {
			return -1
		}

		return r
	}, sanitized)
}

// SanitizeTextSlice applies [SanitizeText] to every element. Empty strings
// after sanitization are preserved (the caller decides whether to filter).
func SanitizeTextSlice(items []string) []string {
	if items == nil {
		return nil
	}

	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, SanitizeText(item))
	}

	return result
}

// SanitizeURL validates a user-supplied URL WITHOUT HTML sanitization and
// returns it with a whitespace trim only.
//
// Why this exists as a separate function: URLs must never run through
// [SanitizeText]. bluemonday parses the value as HTML, so entity-bearing
// query strings are silently rewritten — `&not=` becomes `¬=` and the
// fetch targets the wrong resource (verified live in CV, 2026-09-04).
// Validation requires an absolute http/https URL.
func SanitizeURL(input string) (string, bool) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", false
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host == "" {
		return "", false
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", false
	}

	return trimmed, true
}
