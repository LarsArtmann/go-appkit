// Command security-demo shows the appkit security module's full hardened
// chain on a real appkit service: environment-tuned security headers,
// per-request CSP nonces, keyed rate limiting, origin checks, CSRF with the
// API-key bypass, typed body limits, and shared-secret API-key auth — in
// the canonical wire order (RateLimit → OriginCheck → APIKeyCSRFBypass(CSRF)
// → APIKeyAuth → handler).
//
// Run and try (the demo key activates the guards; an empty key is dev mode,
// where the auth guard passes through):
//
//	API_KEY=demo-key-123 go run ./example
//
//	curl -i http://localhost:8090/ # public page, per-request nonce'd CSP
//	curl -i http://localhost:8090/api/data # 401 — no key
//	curl -i -H 'X-Api-Key: demo-key-123' http://localhost:8090/api/data # 200
//	for i in $(seq 1 12); do curl -s -o /dev/null -w '%{http_code} ' \
//	-H 'X-Api-Key: demo-key-123' http://localhost:8090/api/data; done # ends 429
//	curl -i -X POST -H 'X-Api-Key: demo-key-123' -d '{"msg":"hi"}' \
//	http://localhost:8090/api/echo # 200 — the key skips the CSRF token dance
//	curl -i -X POST -d '{"msg":"hi"}' http://localhost:8090/api/echo # 400 — CSRF
package main

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/security"
	"github.com/larsartmann/httputil"
)

const (
	defaultAddr = ":8090"

	// demoLimit/burst keep the curl demo short: six quick requests trip it.
	demoLimit   uint = 5
	demoBurst   uint = 8
	demoMaxKeys uint = 1_000
	demoBodyMax      = 1 << 20
)

func main() {
	logger := slog.Default()

	cfg := appkit.DefaultServiceConfig()
	cfg.Addr = addrFromEnv()
	cfg.OuterMiddlewares = []appkit.Middleware{
		security.SecurityHeaders(security.HeadersConfig{
			Environment:           security.Production,
			HSTS:                  "",
			ContentSecurityPolicy: "",
		}),
	}

	svc, err := appkit.NewService(cfg)
	if err != nil {
		logger.Error("create service", "error", err)

		return
	}

	apiKey := os.Getenv("API_KEY")

	svc.Mux.HandleFunc("GET /{$}", pageHandler)

	apiChain := hardenedChain(apiKey, logger)
	svc.Mux.Handle("GET /api/data", apiChain(http.HandlerFunc(dataHandler)))
	svc.Mux.Handle("POST /api/echo", apiChain(security.BodyLimit(demoBodyMax)(http.HandlerFunc(echoHandler))))

	if err := svc.Run(cfg.Context()); err != nil { //nolint:contextcheck // demo blocks on the service lifecycle
		logger.Error("service stopped", "error", err)
	}
}

// hardenedChain wires the batteries in the README's canonical order. Each
// route group builds its own chain so limiters never share bucket sets.
func hardenedChain(apiKey string, logger *slog.Logger) func(http.Handler) http.Handler {
	rateLimit := security.RateLimit(security.RateLimitConfig{
		Limit:        demoLimit,
		Burst:        demoBurst,
		Window:       0,
		MaxKeys:      demoMaxKeys,
		KeyExtractor: nil,
	})
	origin := security.OriginCheck([]string{demoOrigin}, logger)
	csrf := security.CSRF(httputil.CSRFConfig{TrustedOrigins: []string{demoOrigin}}, logger)
	auth := security.APIKeyAuth(apiKey)

	return func(h http.Handler) http.Handler {
		return rateLimit(origin(security.APIKeyCSRFBypass(csrf)(auth(h))))
	}
}

// pageHandler serves the public page with a fresh nonce per request: the
// nonce is minted, folded into the CSP via BuildCSP, and echoed into the
// inline script tag — the only way a script may execute under the policy.
func pageHandler(w http.ResponseWriter, r *http.Request) {
	nonce, err := security.GenerateNonce()
	if err != nil {
		http.Error(w, "nonce generation failed", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Security-Policy", security.BuildCSP(security.CSPConfig{
		Environment: security.Production,
		Nonce:       nonce,
		StyleInline: false,
		ConnectSrc:  nil,
		ImgSrc:      nil,
		JSONLD:      false,
	}))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, _ = w.Write([]byte(
		`<!doctype html><html><head><script nonce="` + nonce + `">document.body.dataset.booted="1"</script></head>` +
			`<body><h1>security-demo</h1><p>See the command doc comment for curl walkthroughs.</p></body></html>`))
}

func dataHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]string{"ok": "true", "route": "/api/data"})
}

// echoHandler echoes the request body; the route wraps it in a typed body
// limit, so oversize bodies surface as *http.MaxBytesError (413), never a
// silent truncation.
func echoHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			http.Error(w, "body too large", http.StatusRequestEntityTooLarge)

			return
		}

		http.Error(w, "body read failed", http.StatusBadRequest)

		return
	}

	writeJSON(w, map[string]string{"echoed": string(body)})
}

func writeJSON(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

func addrFromEnv() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}

	return defaultAddr
}
