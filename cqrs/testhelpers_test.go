package cqrs

import (
	"context"
	"os"
	"testing"

	"github.com/larsartmann/go-cqrs-lite/system/v4"
)

// memoryDeployment returns a minimal operator config with a memory engine
// and the two canonical instances.
func memoryDeployment() *system.DeploymentConfig {
	return &system.DeploymentConfig{
		Engines: map[string]system.EngineConfig{
			defaultEngineName: {Driver: memoryDriver},
		},
		Instances: []system.InstanceConfig{
			{Role: system.RoleSourceOfTruth, Engine: defaultEngineName},
			{Role: system.RoleProjections, Engine: defaultEngineName},
		},
	}
}

// writeFile writes content to path with 0o600 permissions.
func writeFile(path, content string) error {
	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		return err //nolint:wrapcheck // test helper
	}

	return nil
}

// newTestEventService starts an EventService backed by a throwaway SQLite
// database in a per-test temp dir (unless cfg.DSN is already set) and
// registers its shutdown as test cleanup, replacing the hand-rolled
// New/err-check/defer-shutdown preamble that used to open most tests.
func newTestEventService(t *testing.T, cfg EventConfig) *EventService {
	t.Helper()

	if cfg.DSN == "" {
		cfg.DSN = t.TempDir() + "/test.db"
	}

	svc, err := NewEventService(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Cleanup(func() { _ = svc.Shutdown(context.Background()) })

	return svc
}
