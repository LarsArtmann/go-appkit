// Package httpxtest pins the httpx wrapper contracts from test code: the
// ResponseWriter-interface forwarding guarantee (B3) and, as the module
// grows, the other testable contracts of the handler-DX battery.
//
// It is a test-only dependency by design — importing it from production
// code is a category error the package name warns about.
package httpxtest

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"
)

// probeWriter is a ResponseWriter that supports every optional interface
// the wrapper contract covers and records whether calls actually arrived:
// a wrapper that satisfies an assertion with a no-op lie (claims Flusher,
// swallows the flush) fails the forwarding check, not just the capability
// check.
type probeWriter struct {
	header             http.Header
	writes             []byte
	status             int
	flushes            int
	hijacked           bool
	wroteHeader        bool
	fullDuplexAttempts int
}

// errNotAConnection is the probe's EnableFullDuplex answer: the probe is
// not a real connection, so the capability always "fails" — what the
// findings check is that the ATTEMPT reached the base, not that it
// succeeded.
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

// Hijack matches http.Hijacker exactly: the probe "fails" the hijack
// (there is no connection), but records the ATTEMPT — what the findings
// check is that the wrapper forwarded the call, not that it succeeded.
func (w *probeWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true

	return nil, nil, errNotAConnection //nolint:wrapcheck // test probe
}

func (w *probeWriter) EnableFullDuplex() error {
	w.fullDuplexAttempts++

	return errNotAConnection //nolint:wrapcheck // test probe
}

// AssertForwardsInterfaces fails the test when a ResponseWriter wrapper
// hides an optional interface its base supports, or forwards it in name
// only. Pass a factory that wraps a given base writer — typically the
// constructor of the wrapper under test:
//
//	httpxtest.AssertForwardsInterfaces(t, func(base http.ResponseWriter) http.ResponseWriter {
//		return myWrapper{base}
//	})
//
// The helper builds a probe base that implements http.Flusher,
// http.Hijacker, and full-duplex control, applies wrap, and reports every
// broken forwarding: hidden interfaces, swallowed writes and statuses,
// no-op stubs that only satisfy the type assertion, and an
// http.ResponseController that cannot reach the base.
//
// The factory shape (not a pre-built wrapper) is what makes the check
// sound: the helper controls the base, so it knows exactly which optional
// interfaces the wrapper must not hide. This is the discrimination the
// battery was cut for — the plain-embedding wrapper shape (the
// notFoundInterceptor class) compiles clean, passes Recorder-based tests,
// and fails here.
func AssertForwardsInterfaces(t testing.TB, wrap func(base http.ResponseWriter) http.ResponseWriter) {
	t.Helper()

	for _, finding := range ForwardingFindings(wrap) {
		t.Error(finding)
	}
}

// ForwardingFindings is AssertForwardsInterfaces without the testing
// dependency: it returns the wrapper's forwarding violations as strings
// (empty means the contract holds). Exposed for callers that want to
// assert the discriminator itself or report findings through their own
// channels.
func ForwardingFindings(wrap func(base http.ResponseWriter) http.ResponseWriter) []string {
	if wrap == nil {
		return []string{"httpxtest: wrap factory is nil"}
	}

	base := newProbeWriter()
	wrapped := wrap(base)

	var findings []string

	if _, ok := wrapped.(http.Flusher); !ok {
		findings = append(findings, describe(wrapped)+" hides http.Flusher — SSE and streaming handlers through it stall in production")
	}

	if _, ok := wrapped.(http.Hijacker); !ok {
		findings = append(findings, describe(wrapped)+" hides http.Hijacker — connection upgrades through it fail at runtime")
	}

	wrapped.Header().Set("X-Probe", "1")
	if base.header.Get("X-Probe") != "1" {
		findings = append(findings, describe(wrapped)+" does not share the header map with the base")
	}

	if _, err := wrapped.Write([]byte("body")); err != nil {
		findings = append(findings, describe(wrapped)+" write failed: "+err.Error())
	} else if len(base.writes) == 0 {
		findings = append(findings, describe(wrapped)+" swallowed the body write — the base never saw the bytes")
	}

	wrapped.WriteHeader(http.StatusTeapot)
	if !base.wroteHeader || base.status != http.StatusTeapot {
		findings = append(findings, describe(wrapped)+" swallowed WriteHeader (the base never saw the status)")
	}

	if flusher, ok := wrapped.(http.Flusher); ok {
		flusher.Flush()
		if base.flushes == 0 {
			findings = append(findings, describe(wrapped)+" satisfies http.Flusher but does not forward Flush — a no-op stub that only passes the type assertion")
		}
	}

	if hijacker, ok := wrapped.(http.Hijacker); ok {
		_, _, _ = hijacker.Hijack()
		if !base.hijacked {
			findings = append(findings, describe(wrapped)+" satisfies http.Hijacker but does not forward Hijack")
		}
	}

	rc := http.NewResponseController(wrapped)
	_ = rc.EnableFullDuplex()
	if base.fullDuplexAttempts == 0 {
		findings = append(findings, describe(wrapped)+" is opaque to http.ResponseController — neither Unwrap nor a full-duplex capability lets the controller reach the base")
	}

	return findings
}

func describe(w http.ResponseWriter) string {
	return fmt.Sprintf("wrapper %T", w)
}
