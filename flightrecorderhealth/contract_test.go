package flightrecorderhealth_test

import (
	frhealth "github.com/larsartmann/go-appkit/flightrecorderhealth"
	"github.com/larsartmann/go-health"
	"github.com/samber/do/v2"
)

// Compile-time proof that Trigger satisfies health.HealthRecorder — this
// module's core contract. If go-health ever changes the interface, this
// assertion becomes a compile error instead of a silent runtime break.
var _ health.HealthRecorder = (*frhealth.Trigger)(nil)

// Compile-time proof that Checkable satisfies do.HealthcheckerWithContext, so
// the recorder's operational state is discoverable by the health dashboard.
var _ do.HealthcheckerWithContext = (*frhealth.Checkable)(nil)
