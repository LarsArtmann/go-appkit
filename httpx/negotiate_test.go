package httpx

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larsartmann/go-codec"
)

func negotiateRequest(t *testing.T, accept string) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/thing", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if accept != "" {
		req.Header.Set("Accept", accept)
	} else {
		req.Header.Del("Accept")
	}

	return req
}

func TestNegotiate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		accept string
		offers []string
		want   string
	}{
		{"absent header serves first offer", "", []string{"text/html", "application/json"}, "text/html"},
		{"exact match", "application/json", []string{"text/html", "application/json"}, "application/json"},
		{
			"q ranking",
			"text/html;q=0.3, application/json;q=0.9",
			[]string{"text/html", "application/json"},
			"application/json",
		},
		{
			"tie keeps offer order",
			"text/html, application/json",
			[]string{"text/html", "application/json"},
			"text/html",
		},
		{
			"reordered offers keep preference order",
			"application/json, text/html",
			[]string{"text/html", "application/json"},
			"text/html",
		},
		{"full wildcard serves first offer", "*/*", []string{"text/html", "application/json"}, "text/html"},
		{"subtype wildcard beats unmentioned", "text/*", []string{"application/json", "text/html"}, "text/html"},
		{
			"subtype wildcard loses on q",
			"text/*;q=0.8, application/json;q=0.9",
			[]string{"text/html", "application/json"},
			"application/json",
		},
		{
			"parameters ignored in matching",
			"TEXT/HTML; charset=utf-8",
			[]string{"text/html", "application/json"},
			"text/html",
		},
		{
			"q zero refuses the offer",
			"text/html;q=0, */*",
			[]string{"text/html", "application/json"},
			"application/json",
		},
		{"all offers refused", "text/html;q=0", []string{"text/html"}, ""},
		{"unrelated types only", "image/png", []string{"text/html", "application/json"}, ""},
		{"empty offers", "text/html", nil, ""},
		{
			"malformed q falls back to default 1",
			"text/html;q=banana, application/json;q=0.5",
			[]string{"text/html", "application/json"},
			"text/html",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := Negotiate(negotiateRequest(t, tc.accept), tc.offers...); got != tc.want {
				t.Fatalf("Negotiate(%q, %v) = %q, want %q", tc.accept, tc.offers, got, tc.want)
			}
		})
	}
}

func TestRequestCodec(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		contentType string
		wantEnc     codec.Encoding
		wantErr     bool
	}{
		{"empty defaults to JSON", "", codec.EncodingJSON, false},
		{"json", "application/json", codec.EncodingJSON, false},
		{"cbor", "application/cbor", codec.EncodingCBOR, false},
		{"json with charset", "application/json; charset=utf-8", codec.EncodingJSON, false},
		{"structured suffix +json", "application/vnd.acme+json", codec.EncodingJSON, false},
		{"structured suffix +cbor", "application/vnd.acme+cbor", codec.EncodingCBOR, false},
		{"uppercase type", "Application/CBOR", codec.EncodingCBOR, false},
		{"unsupported", "text/plain", "", true},
		{"unsupported vendored", "application/xml", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/thing", strings.NewReader("{}"))
			if err != nil {
				t.Fatalf("build request: %v", err)
			}

			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			got, err := RequestCodec(req)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("RequestCodec(%q) = %v, want an error", tc.contentType, got)
				}

				return
			}

			if err != nil {
				t.Fatalf("RequestCodec(%q): %v", tc.contentType, err)
			}

			if got.Encoding() != tc.wantEnc {
				t.Fatalf("RequestCodec(%q).Encoding() = %q, want %q", tc.contentType, got.Encoding(), tc.wantEnc)
			}
		})
	}
}

func TestRequestCodec_UnsupportedIs415Rejection(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/thing", strings.NewReader("x"))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	req.Header.Set("Content-Type", "text/csv")

	_, err = RequestCodec(req)
	if err == nil {
		t.Fatal("want an error for text/csv")
	}

	rec := httptest.NewRecorder()
	WriteError(rec, req, err)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("WriteError status = %d, want 415", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"family":"rejection"`) {
		t.Fatalf("body = %s, want the rejection family rendered", rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), "negotiate.unsupported_media_type") {
		t.Fatalf("body = %s, want the negotiate code rendered", rec.Body.String())
	}
}

func TestResponseCodec(t *testing.T) {
	t.Parallel()

	jsonCodec, _ := codec.ForEncoding(codec.EncodingJSON)
	cborCodec, _ := codec.ForEncoding(codec.EncodingCBOR)

	cases := []struct {
		name    string
		accept  string
		reqEnc  codec.Encoding
		wantEnc codec.Encoding
	}{
		{"absent falls back to request codec", "", codec.EncodingCBOR, codec.EncodingCBOR},
		{"wildcard falls back to request codec", "*/*", codec.EncodingJSON, codec.EncodingJSON},
		{"cbor preferred", "application/cbor", codec.EncodingJSON, codec.EncodingCBOR},
		{"json preferred", "application/json", codec.EncodingCBOR, codec.EncodingJSON},
		{"q prefers json", "application/json;q=0.9, application/cbor;q=0.5", codec.EncodingJSON, codec.EncodingJSON},
		{"q prefers cbor", "application/json;q=0.5, application/cbor;q=0.9", codec.EncodingJSON, codec.EncodingCBOR},
		{"tie prefers cbor", "application/json, application/cbor", codec.EncodingJSON, codec.EncodingCBOR},
		{"refused json substituted", "application/json;q=0", codec.EncodingJSON, codec.EncodingCBOR},
		{"refused cbor substituted", "application/cbor;q=0", codec.EncodingCBOR, codec.EncodingJSON},
		{"refused other keeps request codec", "application/json;q=0", codec.EncodingCBOR, codec.EncodingCBOR},
		{"unrelated keeps request codec", "text/html", codec.EncodingJSON, codec.EncodingJSON},
		{"suffix counts as cbor", "application/vnd.acme+cbor", codec.EncodingJSON, codec.EncodingCBOR},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reqCodec := jsonCodec
			if tc.reqEnc == codec.EncodingCBOR {
				reqCodec = cborCodec
			}

			got := ResponseCodec(negotiateRequest(t, tc.accept), reqCodec)
			if got.Encoding() != tc.wantEnc {
				t.Fatalf("ResponseCodec(%q).Encoding() = %q, want %q", tc.accept, got.Encoding(), tc.wantEnc)
			}
		})
	}
}

// The end-to-end codec seam: decode a CBOR request, encode the response
// through Accept-negotiated CBOR — the round-trip a JSON+CBOR API needs.
func TestCodecSeam_CBORRoundTrip(t *testing.T) {
	t.Parallel()

	payload := map[string]string{"name": "thing"}

	encoded, err := codecForEncoding(codec.EncodingCBOR).Encode(payload)
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/thing", bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	req.Header.Set("Content-Type", "application/cbor")
	req.Header.Set("Accept", "application/cbor")

	reqCodec, err := RequestCodec(req)
	if err != nil {
		t.Fatalf("RequestCodec: %v", err)
	}

	if reqCodec.Encoding() != codec.EncodingCBOR {
		t.Fatalf("request codec = %q, want cbor", reqCodec.Encoding())
	}

	var decoded map[string]string
	if err := reqCodec.Decode(encoded, &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if decoded["name"] != "thing" {
		t.Fatalf("decoded = %v, want the payload", decoded)
	}

	respCodec := ResponseCodec(req, reqCodec)
	if respCodec.Encoding() != codec.EncodingCBOR {
		t.Fatalf("response codec = %q, want cbor", respCodec.Encoding())
	}

	out, err := respCodec.Encode(decoded)
	if err != nil {
		t.Fatalf("encode response: %v", err)
	}

	var back map[string]string
	if err := codecForEncoding(codec.EncodingCBOR).Decode(out, &back); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if back["name"] != "thing" {
		t.Fatalf("round-trip lost the payload: %v", back)
	}
}
