package cqrs

import (
	"context"

	"github.com/larsartmann/go-cqrs-lite/system/v4"
)

// HealthCheck verifies that all infrastructure resources are reachable and
// healthy: the system is not stopped, every engine implementing
// metaengine.HealthChecker answers a ping, and no projection worker is in a
// failed state. Returns the first error otherwise. Use this for Kubernetes
// liveness/readiness probes; pair with LagPerProjection for read-your-writes
// gating (delegates to system.System.HealthCheck).
func (es *EventService) HealthCheck(ctx context.Context) error {
	return es.sys.HealthCheck(ctx) //nolint:wrapcheck // delegation
}

// EngineHealth reports the health status of every engine that implements
// metaengine.HealthChecker, plus projection host worker status. Unlike
// HealthCheck, which returns only the first failure, this reports ALL
// engines — suitable for detailed dashboards and debugging
// (delegates to system.System.HealthCheckDetailed).
func (es *EventService) EngineHealth(ctx context.Context) []system.EngineHealth {
	return es.sys.HealthCheckDetailed(ctx)
}

// ScreamReport returns the system's safety report: construction-time config
// findings plus any plan-drift findings from the pinned projection manifest.
// Always non-nil; iterate Diagnostics for severity. NewEventService already
// logs warnings and errors from the construction-time report — this accessor
// exists for dashboards and tests that want the structured findings.
func (es *EventService) ScreamReport() *system.ScreamReport {
	return es.sys.ScreamReport()
}
