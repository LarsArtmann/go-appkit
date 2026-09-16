package health_test

import (
	gohealth "github.com/larsartmann/go-health"
	dashboard "github.com/larsartmann/go-health-dashboard"
)

// Compile-time contract assertion: the dashboard reads the probe through the
// structural [dashboard.Prober] interface, and this module's whole value is
// bridging a [gohealth.Probe] into it. If either dependency changes that
// surface in a release bump, the build breaks HERE instead of failing
// silently at runtime — the flightrecorderhealth-module pattern (contract
// drift is a compile error, not a release surprise).
var _ dashboard.Prober = (*gohealth.Probe)(nil)
