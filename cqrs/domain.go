package cqrs

import (
	"slices"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/projectionhost/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
)

// hostBootstrapDeclaration names the zero-entry count projection that
// guarantees system.New creates the projection host even when consumers only
// register raw host projections (the host is built exclusively when
// DomainConfig.Projections is non-empty). It counters no event types and has
// no read-model cost.
const hostBootstrapDeclaration = "appkit-host"

// bootstrapSample is the decoder sample for the never-emitted bootstrap
// event type.
type bootstrapSample struct {
	ID string
}

// bootstrapProjections returns the host-guaranteeing declaration used when no
// consumer Domain declares projections of its own.
func bootstrapProjections() []system.ProjectionDeclaration {
	return []system.ProjectionDeclaration{
		system.Count(hostBootstrapDeclaration).
			On("appkit.internal.never", bootstrapSample{ID: ""}, 1, "ID").
			Done(),
	}
}

// mergeDomain folds the consumer's EventConfig.Domain into the
// wrapper-derived system.DomainConfig handed to system.New.
//
// Merge contract (pinned by domain_merge_test.go):
//
//   - Middleware: [in-flight drain tracker] + Domain.Middleware +
//     CommandMiddleware — the tracker stays outermost so drain times the
//     whole chain.
//   - Projections: the wrapper's host-bootstrap declaration is appended ONLY
//     when the consumer declares none (system.New builds the projection host
//     exclusively when DomainConfig.Projections is non-empty).
//   - ProjectionHostOptions: Domain.ProjectionHostOptions, then
//     EventConfig.HostOptions, then the wrapper's derived wiring — derived
//     wiring wins conflicts.
//   - CheckpointStore: EventConfig.CheckpointStore (the wrapper's knob) wins;
//     otherwise Domain.CheckpointStore applies; nil means system's
//     engine-backed default (ADR-0142).
//   - Everything else (Commands, Queries, Timers, Events,
//     DisableCoeffectValidation, Evolutions, ProjectionDecoder,
//     ProjectionTypeDecoder, ProjectionEventDecoder, ShutdownDependencies)
//     passes through verbatim.
//
// A nil Domain yields the pre-seam wrapper behavior byte-for-byte.
func mergeDomain(
	cfg EventConfig,
	inFile *inFlightTracker,
	dlqStore projectionhost.DeadLetterStore,
) system.DomainConfig {
	derivedHostOptions := cfg.hostOptions(dlqStore)

	if cfg.Domain == nil {
		return system.DomainConfig{ //nolint:exhaustruct_v5 // domain registration is the consumer's job
			Middleware:            append([]command.Middleware{inFile.commandMiddleware()}, cfg.CommandMiddleware...),
			Projections:           bootstrapProjections(),
			ProjectionHostOptions: derivedHostOptions,
			CheckpointStore:       cfg.CheckpointStore,
		}
	}

	domain := *cfg.Domain // shallow copy: declarations are immutable once built

	domain.Middleware = concatCommandMiddleware(
		[]command.Middleware{inFile.commandMiddleware()}, domain.Middleware, cfg.CommandMiddleware)
	domain.ProjectionHostOptions = append(
		slices.Clone(domain.ProjectionHostOptions), derivedHostOptions...)

	if len(domain.Projections) == 0 {
		domain.Projections = bootstrapProjections()
	}

	if cfg.CheckpointStore != nil {
		domain.CheckpointStore = cfg.CheckpointStore
	}

	return domain
}

// concatCommandMiddleware concatenates middleware groups into a fresh slice,
// so no caller-owned backing array is ever written through (mirrors the core
// module's concatMiddlewares invariant).
func concatCommandMiddleware(groups ...[]command.Middleware) []command.Middleware {
	var total int
	for _, group := range groups {
		total += len(group)
	}

	merged := make([]command.Middleware, 0, total)
	for _, group := range groups {
		merged = append(merged, group...)
	}

	return merged
}
