package flightrecorderhealth_test

import (
	"context"
	"io"
	"testing"
	"time"

	frhealth "github.com/larsartmann/go-appkit/flightrecorderhealth"
	fr "github.com/larsartmann/go-flightrecorder"
	"github.com/samber/do/v2"
)

// BenchmarkTrigger_RecordHealthCheckWithContext_AllPass measures the hot path:
// a fully passing health-check batch flows through the Trigger without
// capturing (trigger evaluates false, no snapshot I/O). The recorder is
// deliberately not started — the no-capture path never touches it beyond the
// nil check, so this isolates the Trigger's own overhead.
func BenchmarkTrigger_RecordHealthCheckWithContext_AllPass(b *testing.B) {
	rec, err := fr.New(fr.WithWriter(io.Discard), fr.WithMinAge(50*time.Millisecond))
	if err != nil {
		b.Fatalf("fr.New() error: %v", err)
	}

	injector := do.New()
	registerSvc(injector, "database", nil)
	registerSvc(injector, "cache", nil)

	trigger := frhealth.NewTrigger(rec, frhealth.WithCooldown(time.Second))

	ctx := context.Background()

	b.ResetTimer()

	for b.Loop() {
		_ = trigger.RecordHealthCheckWithContext(ctx, injector)
	}
}
