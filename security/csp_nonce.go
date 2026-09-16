package security

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"slices"
	"strings"
)

// nonceBytes is the number of random bytes used to generate a CSP nonce.
// 16 bytes (128 bits) is the OWASP-recommended minimum.
const nonceBytes = 16

// scriptSrcDirective is the CSP directive controlling script execution.
const scriptSrcDirective = "script-src"

// GenerateNonce generates a cryptographically random nonce suitable for CSP.
// Returns a base64-encoded string.
func GenerateNonce() (string, error) {
	b := make([]byte, nonceBytes)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate CSP nonce: %w", err)
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

// cspNonceKey is a typed context key for storing the CSP nonce.
type cspNonceKey struct{}

// WithNonce stores a CSP nonce in the context.
func WithNonce(ctx context.Context, nonce string) context.Context {
	return context.WithValue(ctx, cspNonceKey{}, nonce)
}

// NonceFromContext retrieves the CSP nonce from the context. Returns "" if
// no nonce is present. Component libraries consume this extractor contract
// (go-health-dashboard takes it via WithNonceExtractor).
func NonceFromContext(ctx context.Context) string {
	nonce, _ := ctx.Value(cspNonceKey{}).(string)

	return nonce
}

// AddNonceToScriptSrc parses a CSP header string and adds 'nonce-{nonce}'
// to the script-src directive's source list. If script-src is not present,
// the CSP is returned unchanged; an already-present identical nonce is not
// duplicated. The result is normalized to the "; " (semicolon + space)
// directive separator [BuildCSP] emits.
func AddNonceToScriptSrc(csp, nonce string) string {
	raw := strings.Split(csp, ";")

	directives := make([]string, 0, len(raw))
	for _, directive := range raw {
		if trimmed := strings.TrimSpace(directive); trimmed != "" {
			directives = append(directives, trimmed)
		}
	}

	for i, directive := range directives {
		parts := strings.Fields(directive)
		if len(parts) == 0 || parts[0] != scriptSrcDirective {
			continue
		}

		nonceToken := fmt.Sprintf("'nonce-%s'", nonce)

		if slices.Contains(parts, nonceToken) {
			return strings.Join(directives, "; ")
		}

		parts = append(parts, nonceToken)
		directives[i] = strings.Join(parts, " ")

		break
	}

	return strings.Join(directives, "; ")
}
