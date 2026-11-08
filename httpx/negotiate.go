package httpx

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/larsartmann/go-codec"
	errorfamily "github.com/larsartmann/go-error-family"
)

// Negotiate picks the best media type from offers for the request's
// Accept header (RFC 9110 §12.5.1) — the one-line answer for handlers
// that serve the same resource in several representations:
//
//	switch httpx.Negotiate(r, "text/html", "application/json") {
//	case "text/html":
//		renderPage(w, r)
//	default:
//		renderJSON(w, r)
//	}
//
// Ranking: q-value first, then match specificity (exact type beats
// subtype wildcard beats */*), then the offers' own order — the first
// offer is the server's preference and wins every tie. An absent Accept
// header accepts anything (first offer). A q of 0 is an explicit
// refusal: that offer can never win. When the header mentions none of
// the offers and carries no wildcard, no representation is acceptable —
// Negotiate returns "" and the caller answers 406 (or serves its
// default; both are defensible, the battery only makes the situation
// visible).
//
// Media types compare on their bare, lowercased form; parameters
// (charset, q, vendor specifics) affect only the quality value.
func Negotiate(r *http.Request, offers ...string) string {
	if len(offers) == 0 {
		return ""
	}

	accept := r.Header.Get("Accept")

	best := negotiationCandidate{index: -1}

	for i, offer := range offers {
		c := negotiationCandidate{
			index:       i,
			quality:     0,
			specificity: -1,
		}

		if q, s, mentioned := negotiateMatch(bareMediaType(offer), accept); mentioned {
			if q == 0 {
				continue // explicit refusal can never win
			}

			c.quality = q
			c.specificity = s
		} else if accept != "" {
			// The header names types but none of them (nor any wildcard)
			// accepts this offer — it is not acceptable, not a default.
			continue
		}

		if best.index == -1 || c.beats(best) {
			best = c
		}
	}

	if best.index == -1 {
		return ""
	}

	return offers[best.index]
}

// negotiationCandidate is one offer's standing in the Negotiate ranking.
// A candidate the header never mentioned stays eligible at the bottom
// (specificity -1): it loses to any explicit mention but keeps the
// offers' preference order among its own kind.
type negotiationCandidate struct {
	quality     float64
	specificity int
	index       int
}

func (c negotiationCandidate) beats(other negotiationCandidate) bool {
	if c.quality != other.quality {
		return c.quality > other.quality
	}

	if c.specificity != other.specificity {
		return c.specificity > other.specificity
	}

	return c.index < other.index
}

// negotiateMatch finds how the Accept header values a given bare media
// type: the best (q, specificity) among all entries that accept it. An
// absent header accepts everything at wildcard level (RFC: absent Accept
// = */*). specificity: 2 = exact type, 1 = subtype wildcard (type/*),
// 0 = */*.
func negotiateMatch(bareOffer, accept string) (quality float64, specificity int, mentioned bool) {
	if accept == "" {
		return 1, 0, true // absent header: everything acceptable at wildcard preference
	}

	best := negotiationCandidate{index: -1}

	for entry := range strings.SplitSeq(accept, ",") {
		mediaRange, params, _ := strings.Cut(entry, ";")

		s, ok := rangeMatches(bareOffer, bareMediaType(mediaRange))
		if !ok {
			continue
		}

		q := qualityParam(params)

		// The MOST SPECIFIC matching range governs (RFC: a specific
		// reference takes precedence over a less specific one) — a q=0 on
		// the exact type is a refusal a wildcard cannot resurrect.
		if best.index == -1 || s > best.specificity || (s == best.specificity && q > best.quality) {
			best = negotiationCandidate{quality: q, specificity: s, index: 0}
		}
	}

	if best.index == -1 {
		return 0, 0, false
	}

	return best.quality, best.specificity, true
}

// rangeMatches reports whether an Accept media range accepts the offer's
// bare type, and how specifically.
func rangeMatches(bareOffer, bareRange string) (int, bool) {
	switch {
	case bareRange == "*/*" || bareRange == "*":
		return 0, true
	case strings.HasSuffix(bareRange, "/*"):
		offerType, _, _ := strings.Cut(bareOffer, "/")
		rangeType, _, _ := strings.Cut(bareRange, "/")

		return 1, offerType == rangeType
	default:
		return 2, bareOffer == bareRange
	}
}

// bareMediaType extracts the lowercased media type from a header value,
// stripping parameters: "text/html; q=0.5" → "text/html".
func bareMediaType(headerValue string) string {
	mediaType, _, _ := strings.Cut(headerValue, ";")

	return strings.ToLower(strings.TrimSpace(mediaType))
}

// qualityParam extracts the q parameter (case-insensitive name, RFC 9110
// §5.6.2.1) from a parameter list, defaulting to 1 and clamped to
// [0,1]; malformed values fall back to the default so a parseable q
// elsewhere in the entry still wins.
func qualityParam(params string) float64 {
	for param := range strings.SplitSeq(params, ";") {
		name, value, _ := strings.Cut(param, "=")
		if !strings.EqualFold(strings.TrimSpace(name), "q") {
			continue
		}

		q, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil || q < 0 || q > 1 {
			continue
		}

		return q
	}

	return 1
}

// ErrUnsupportedMediaType is returned by [RequestCodec] when the request
// declares a Content-Type the battery cannot decode. It carries the
// Rejection family with HTTP status 415, so [WriteError] renders it with
// the exact status and the no-leak contract.
var ErrUnsupportedMediaType error = &unsupportedMediaTypeError{}

type unsupportedMediaTypeError struct{}

func (*unsupportedMediaTypeError) Error() string {
	return "httpx: the request's Content-Type is not a supported media type (want JSON or CBOR)"
}

func (*unsupportedMediaTypeError) ErrorFamily() errorfamily.Family { return errorfamily.Rejection }

func (*unsupportedMediaTypeError) ErrorCode() string { return "negotiate.unsupported_media_type" }

func (*unsupportedMediaTypeError) HTTPStatus() int { return http.StatusUnsupportedMediaType }

// RequestCodec resolves the [codec.Codec] for decoding the REQUEST body
// from its Content-Type — the go-codec seam of the negotiation battery:
// application/json (and +json structured suffixes, RFC 6839) decode as
// JSON, application/cbor (+cbor) as CBOR, and an empty Content-Type
// defaults to JSON. Anything else returns ErrUnsupportedMediaType
// (415 through [WriteError]).
//
//	reqCodec, err := httpx.RequestCodec(r)
//	if err != nil {
//		httpx.WriteError(w, r, err)
//
//		return
//	}
func RequestCodec(r *http.Request) (codec.Codec, error) {
	contentType := r.Header.Get("Content-Type")

	switch enc, ok := encodingForContentType(contentType); {
	case ok:
		return codec.ForEncoding(enc) //nolint:wrapcheck // ForEncoding's classified error is the honest return
	default:
		return nil, errorfamily.Wrapf( //nolint:wrapcheck // constructor is the sanctioned terminal form
			ErrUnsupportedMediaType, errorfamily.Rejection,
			"negotiate.unsupported_media_type",
			"Content-Type %q is not decodable (want JSON or CBOR)", contentType,
		).WithHTTPStatus(http.StatusUnsupportedMediaType)
	}
}

// ResponseCodec resolves the [codec.Codec] for encoding the RESPONSE body
// from the Accept header, falling back to reqCodec (typically
// RequestCodec's result) when the client expresses no preference:
// absent header, wildcards, or only unrelated types. With both
// encodings mentioned, the higher q wins and a tie prefers CBOR
// (denser); q=0 is an explicit refusal and the refused encoding is never
// served when the other is available. Responses chosen this way should
// carry "Vary: Accept" so caches key on the negotiation dimension.
//
// Only the media type is negotiated — Accept-Charset and
// Accept-Encoding (transport codings) are out of scope.
func ResponseCodec(r *http.Request, reqCodec codec.Codec) codec.Codec {
	accept := r.Header.Get("Accept")

	cborQ, cborWanted := codecQuality(accept, codec.EncodingCBOR)
	jsonQ, jsonWanted := codecQuality(accept, codec.EncodingJSON)

	jsonCodec := codecForEncoding(codec.EncodingJSON)
	cborCodec := codecForEncoding(codec.EncodingCBOR)

	switch {
	case cborWanted && jsonWanted:
		switch {
		case jsonQ > cborQ:
			return jsonCodec
		case cborQ > jsonQ, cborQ > 0:
			return cborCodec
		default:
			return reqCodec
		}
	case cborWanted:
		if cborQ > 0 {
			return cborCodec
		}

		if reqCodec.Encoding() == codec.EncodingCBOR {
			return jsonCodec
		}

		return reqCodec
	case jsonWanted:
		if jsonQ > 0 {
			return jsonCodec
		}

		if reqCodec.Encoding() == codec.EncodingJSON {
			return cborCodec
		}

		return reqCodec
	default:
		return reqCodec
	}
}

// codecQuality reports the highest q the Accept header assigns to any
// media type representing enc (bare type or structured-suffix variant,
// RFC 6839), and whether such a type is mentioned at all. Wildcards do
// not count as a mention: they express no preference between JSON and
// CBOR, so callers fall back to the request codec when only wildcards
// match (the recorded negotiation contract).
func codecQuality(accept string, enc codec.Encoding) (float64, bool) {
	base, suffix := "application/json", "+json"
	if enc == codec.EncodingCBOR {
		base, suffix = "application/cbor", "+cbor"
	}

	var (
		quality   float64
		mentioned bool
	)

	for entry := range strings.SplitSeq(accept, ",") {
		mediaRange, params, _ := strings.Cut(entry, ";")

		bare := bareMediaType(mediaRange)
		if bare != base && !strings.HasSuffix(bare, suffix) {
			continue
		}

		if q := qualityParam(params); !mentioned || q > quality {
			quality, mentioned = q, true
		}
	}

	return quality, mentioned
}

func codecForEncoding(enc codec.Encoding) codec.Codec {
	c, err := codec.ForEncoding(enc)
	if err != nil {
		// Unreachable for the two built-in encodings above; ForEncoding
		// only fails for unknown encodings, which these are not.
		panic(err)
	}

	return c
}

// encodingForContentType maps a Content-Type value to its codec
// encoding, honoring structured suffixes (RFC 6839): vendored types like
// application/vnd.acme+cbor count as their base type.
func encodingForContentType(contentType string) (codec.Encoding, bool) {
	if contentType == "" {
		return codec.EncodingJSON, true // no declaration: JSON is the battery default
	}

	bare := bareMediaType(contentType)

	switch {
	case bare == "application/cbor" || strings.HasSuffix(bare, "+cbor"):
		return codec.EncodingCBOR, true
	case bare == "application/json" || strings.HasSuffix(bare, "+json"):
		return codec.EncodingJSON, true
	default:
		return "", false
	}
}
