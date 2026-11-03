package appkit_test

import (
	"testing"

	servertiming "github.com/larsartmann/httputil/server_timing"
)

// The Server-Timing middleware rides appkit's default stack (outermost,
// owner ruling 2026-11-03); this benchmark isolates its per-request cost
// with the same harness as the Logging benchmarks — real round-trips,
// response bodies discarded server-side so measured work is middleware
// overhead, never terminal I/O.
func BenchmarkServerTimingMiddleware(b *testing.B) {
	server := benchmarkServer(b, servertiming.ServerTimingMiddleware())

	b.ReportAllocs()

	for b.Loop() {
		requestOnce(b, server)
	}
}
