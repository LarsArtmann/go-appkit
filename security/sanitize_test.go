package security_test

import (
	"testing"

	"github.com/larsartmann/go-appkit/security"
)

func TestSanitizeText_StripsTagsAndControlChars(t *testing.T) {
	t.Parallel()

	if got := security.SanitizeText("  <script>alert(1)</script>Hello\x00World  "); got != "HelloWorld" {
		t.Errorf("SanitizeText = %q, want %q", got, "HelloWorld")
	}
}

func TestSanitizeText_PreservesNewlinesAndTabs(t *testing.T) {
	t.Parallel()

	got := security.SanitizeText("line1\nline2\tcol\x01x")
	if got != "line1\nline2\tcolx" {
		t.Errorf("SanitizeText = %q, want newlines/tabs preserved, other controls stripped", got)
	}
}

func TestSanitizeTextSlice(t *testing.T) {
	t.Parallel()

	got := security.SanitizeTextSlice([]string{"<b>a</b>", "b"})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("SanitizeTextSlice = %v, want [a b]", got)
	}

	if got := security.SanitizeTextSlice(nil); got != nil {
		t.Errorf("SanitizeTextSlice(nil) = %v, want nil", got)
	}
}

// TestSanitizeURL_EntityQueryUnmangled is THE regression: bluemonday parses
// its input as HTML and entity-rewrites query strings (`&not=` becomes
// `¬=`), silently retargeting the URL. URLs must never run through the text
// sanitizer.
func TestSanitizeURL_EntityQueryUnmangled(t *testing.T) {
	t.Parallel()

	raw := "https://example.com/scan?url=https://target.example?a=1&not=2"

	got, ok := security.SanitizeURL(raw)
	if !ok {
		t.Fatal("valid URL rejected")
	}

	if got != raw {
		t.Errorf("SanitizeURL = %q, want the input UNCHANGED (an HTML sanitizer would rewrite &not= to ¬=)", got)
	}
}

func TestSanitizeURL_Validation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  bool
	}{
		{"https://example.com/path", true},
		{"http://example.com", true},
		{"  https://example.com  ", true},
		{"ftp://example.com", false},
		{"javascript:alert(1)", false},
		{"//example.com", false},
		{"not a url", false},
		{"", false},
	}

	for _, tc := range cases {
		_, ok := security.SanitizeURL(tc.input)
		if ok != tc.want {
			t.Errorf("SanitizeURL(%q) ok = %v, want %v", tc.input, ok, tc.want)
		}
	}
}
