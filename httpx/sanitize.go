package httpx

import (
	"regexp"

	errorfamily "github.com/larsartmann/go-error-family"
)

// leakPatterns matches the standing LAN-surface leak classes in raw error
// text: filesystem paths (POSIX + Windows), source locations, and SQL
// statements. Anything matched is replaced before the text reaches a
// client. The set is deliberately small and conservative: the PRIMARY
// no-leak defense is the family gate below (server-side families never
// echo raw text at all), the patterns are the second line for
// client-facing messages.
var leakPatterns = regexp.MustCompile(
	`(?:/(?:home|Users|var|tmp|etc|opt|root|srv)/\S+)` +
		`|(?:[A-Za-z]:\\(?:Users|Windows|Program\S*|temp)\S*)` +
		`|(?:\S+\.go:\d+)` +
		`|(?i)\b(?:select|insert\s+into|update|delete\s+from)\b[^;]{0,200}` +
		`|(?:goroutine\s+\d+)`)

const redactedMarker = "[redacted]"

// SanitizeErrorMessage returns the client-facing message for err:
//
//   - An explicit Public() string wins (the error author vouches for the
//     text; the same opt-out errorpages honors).
//   - An explicit Message() string wins next (the error author supplies a
//     pre-cleaned message).
//   - Client-facing families (Rejection, Conflict) keep the error's own
//     text — "field email is required" IS the API response — with the
//     leak patterns scrubbed.
//   - Every other family (the server-side families: Transient,
//     Corruption, Infrastructure, Orchestration, and anything unclassifiable)
//     gets the family's default message. Raw internal error text (paths,
//     SQL, stack frames, driver messages) never leaves the process.
//
// The function is pure and safe for concurrent use.
func SanitizeErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	type publicMessenger interface{ Public() string }

	if m, ok := err.(publicMessenger); ok {
		if public := m.Public(); public != "" {
			return public
		}
	}

	type cleanMessenger interface{ Message() string }

	if m, ok := err.(cleanMessenger); ok {
		return m.Message()
	}

	family := errorfamily.Classify(err)

	if family == errorfamily.Rejection || family == errorfamily.Conflict {
		return scrub(err.Error())
	}

	return family.DefaultMessage()
}

// scrub replaces leak-pattern matches with the redaction marker.
func scrub(text string) string {
	return leakPatterns.ReplaceAllString(text, redactedMarker)
}
