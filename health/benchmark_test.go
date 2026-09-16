package health_test

import (
	"context"
	"testing"

	"github.com/larsartmann/go-appkit/health"
)

// benchmarkNewProbe measures the injector-free probe's full batch cost at
// N=1/5/20 checks (the frh-module house bar; the parity TODO's numbers).
// The batch is the unit of cost — every check runs concurrently per batch.
func benchmarkNewProbe(b *testing.B, n int) {
	checks := make(map[string]health.CheckFunc, n)
	for i := range n {
		checks[string(rune('a'+i))] = func(context.Context) error { return nil }
	}

	probe := health.NewProbe(checks)
	ctx := context.Background()

	b.ResetTimer()

	for b.Loop() {
		if err := probe.HealthCheck(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewProbe_Batch_N1(b *testing.B)   { benchmarkNewProbe(b, 1) }
func BenchmarkNewProbe_Batch_N5(b *testing.B)   { benchmarkNewProbe(b, 5) }
func BenchmarkNewProbe_Batch_N20(b *testing.B)  { benchmarkNewProbe(b, 20) }
