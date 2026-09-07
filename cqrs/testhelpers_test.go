package cqrs

import (
	"os"

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
