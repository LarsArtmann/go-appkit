// Package httpxtest pins the httpx wrapper contracts from test code: the
// ResponseWriter-interface forwarding guarantee (B3) and, as the module
// grows, the other testable contracts of the handler-DX battery.
//
// It is a test-only dependency by design — importing it from production
// code is a category error the package name warns about.
package httpxtest

import (
	"errors"
	"net/http"
	"testing"
)

// probeWriter is a ResponseWriter that supports every optional interface
// the wrapper contract covers and records whether calls actually arrived:
// a wrapper that satisfies an assertion with a no-op lie (claims Flusher,
// swallows the flush) fails the forwarding check, not just the capability
// check.
type probeWriter struct {
	header            http.Header
	writes            []byte
	status            int
	flushes           int
	hijacked          bool
	wroteHeader       bool
	fullDuplexAttempts int
}

// errNotAConnection is the probe's EnableFullDuplex answer: it is not a
// real connection, so the capability always "fails" — the assertion
// checks that the ATTEMPT reached the base, not that it succeeded.
var errNotAConnection = errors.New("httpxtest: the probe writer is not a real connection")

func newProbeWriter() *probeWriter {
	return &probeWriter{header: http.Header{}} //nolint:exhaustruct_v5 // counters start at zero by design
}

func (w *probeWriter) Header() http.Header { return w.header }

func (w *probeWriter) Write(b []byte) (int, error) {
	w.writes = append(w.writes, b...)

	return len(b), nil //nolint:wrapcheck // test probe
}

func (w *probeWriter) WriteHeader(status int) {
	w.status = status
	w.wroteHeader = true
}

func (w *probeWriter) Flush() { w.flushes++ }

func (w *probeWriter) Hijack() { w.hijacked = true }

// AssertForwardsInterfaces fails the test when a ResponseWriter wrapper
// hides an optional interface its base supports, or forwards it in name
// only. Pass a factory that wraps a given base writer — typically the
// constructor of the wrapper under test:
//
//	httpxtest.AssertForwardsInterfaces(t, func(base http.ResponseWriter) http.ResponseWriter {
//		return myWrapper{base}
//	})
//
// The helper builds a probe base that implements http.Flusher and
// http.Hijacker, applies wrap, and asserts the result:
//
//   - implements http.Flusher and http.Hijacker,
//   - actually forwards Flush and Hijack calls to the base (a no-op stub
//     that only satisfies the type assertion fails),
//   - forwards Write and WriteHeader bytes and status,
//   - is reachable from http.ResponseController: the wrapper either
//     implements Unwrap or the controller finds the capability on the
//     wrapper itself.
//
// The factory shape (not a pre-built wrapper) is what makes the check
// sound: the helper controls the base, so it knows exactly which optional
// interfaces the wrapper must not hide. This is the discrimination the
// battery was cut for — the plain-embedding wrapper shape (the
// notFoundInterceptor class) compiles clean, passes Recorder-based tests,
// and fails here.
func AssertForwardsInterfaces(t testing.TB, wrap func(base http.ResponseWriter) http.ResponseWriter) {
	t.Helper()

	if wrap == nil {
		t.Fatal("httpxtest: wrap factory is nil")
	}

	base := newProbeWriter()
	wrapped := wrap(base)

	if _, ok := wrapped.(http.Flusher); !ok {
		t.Errorf("wrapper %T hides http.Flusher — SSE and streaming handlers through it stall in production", wrapped)
	}

	if _, ok := wrapped.(http.Hijacker); !ok {
		t.Errorf("wrapper %T hides http.Hijacker — WebSocket-style upgrades through it fail at runtime", wrapped)
	}

	wrapped.Header().Set("X-Probe", "1")
	if base.header.Get("X-Probe") != "1" {
		t.Errorf("wrapper %T does not share the header map with the base", wrapped)
	}

	if _, err := wrapped.Write([]byte("body")); err != nil {
		t.Fatalf("write through wrapper: %v", err)
	}

	wrapped.WriteHeader(http.StatusTeapot)
	if !base.wroteHeader || base.status != http.StatusTeapot {
		t.Errorf("wrapper %T swallowed WriteHeader (base saw status %d, wroteHeader=%v)", wrapped, base.status, base.wroteHeader)
	}

	if len(base.writes) == 0 {
		t.Errorf("wrapper %T swallowed the body write — the base never saw the bytes", wrapped)
	}

	if flusher, ok := wrapped.(http.Flusher); ok {
		flusher.Flush()
		if base.flushes == 0 {
			t.Errorf("wrapper %T satisfies http.Flusher but does not forward Flush — a no-op stub that only passes the type assertion", wrapped)
		}
	}

	if hijacker, ok := wrapped.(http.Hijacker); ok {
		hijacker.Hijack()
		if !base.hijacked {
			t.Errorf("wrapper %T satisfies http.Hijacker but does not forward Hijack", wrapped)
		}
	}

	rc := http.NewResponseController(wrapped)
	if err := rc.EnableFullDuplex(); err == nil && base.fullDuplexAttempts == 0 {
		t.Errorf("wrapper %T answered EnableFullDuplex without reaching the base — the controller is talking to a stranger", wrapped)
	}
}

// The probe's EnableFullDuplex always errors (it is not a real
// connection); the assertion above checks the ATTEMPT reached the base,
// not success.
func (w *probeWriter) EnableFullDuplex() error {
	w.fullDuplexAttempts++

	return errNotAConnection
}
