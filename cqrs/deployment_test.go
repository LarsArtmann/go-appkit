package cqrs

import (
	"testing"

	"github.com/larsartmann/go-cqrs-lite/system/v4"
)

// The README's "Deployment shapes" YAML snippets live as testdata files so
// they are round-trip-verified against system.LoadConfig — the same parser
// EventConfig.ConfigPath uses. A snippet that stops parsing fails here, not
// in an operator's editor.
func TestDeploymentShapeYAML_RoundTripsThroughLoadConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		file string
	}{
		{name: "buses and publish fan-out", file: "deployment-buses.yaml"},
		{name: "priority, materialized views, durability, mixed pools", file: "deployment-priority-views.yaml"},
		{name: "manifest pinning, acknowledged warnings, cache tier", file: "deployment-manifest.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg, err := system.LoadConfig("testdata/" + tt.file)
			if err != nil {
				t.Fatalf("LoadConfig: %v", err)
			}

			if len(cfg.Engines) == 0 {
				t.Fatal("expected engines to parse")
			}

			if len(cfg.Instances) == 0 {
				t.Fatal("expected instances to parse")
			}
		})
	}
}

func TestDeploymentShapeYAML_BusesAndPublishFanOut(t *testing.T) {
	t.Parallel()

	cfg, err := system.LoadConfig("testdata/deployment-buses.yaml")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if len(cfg.Buses) != 2 {
		t.Fatalf("expected 2 buses, got %d", len(cfg.Buses))
	}

	if cfg.Buses["edge"].Driver != "nats" || cfg.Buses["edge"].URL != "nats://localhost:4222" {
		t.Errorf("edge bus did not parse: %+v", cfg.Buses["edge"])
	}

	var sot *system.InstanceConfig

	for i := range cfg.Instances {
		if cfg.Instances[i].Role == system.RoleSourceOfTruth {
			sot = &cfg.Instances[i]
		}
	}

	if sot == nil {
		t.Fatal("expected a source-of-truth instance")
	}

	if len(sot.Publish) != 2 {
		t.Errorf("expected publish to fan out to 2 buses, got %v", sot.Publish)
	}
}

func TestDeploymentShapeYAML_PriorityViewsDurabilityPools(t *testing.T) {
	t.Parallel()

	cfg, err := system.LoadConfig("testdata/deployment-priority-views.yaml")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	hot := cfg.Engines["hot"]
	if hot.Priority != "ReadSpeed" {
		t.Errorf("expected hot engine priority ReadSpeed, got %q", hot.Priority)
	}

	if len(hot.MaterializedViews) != 1 {
		t.Fatalf("expected 1 materialized view, got %d", len(hot.MaterializedViews))
	}

	view := hot.MaterializedViews[0]
	if view.Collection != "orders" || view.Fn != "SUM" || view.Column != "amount" || view.GroupBy != "customer" {
		t.Errorf("materialized view did not parse: %+v", view)
	}

	if cfg.Priority == nil || cfg.Priority.PerQuery["orders_by_customer"] != "ReadSpeed" {
		t.Errorf("expected per-query priority to parse, got %+v", cfg.Priority)
	}

	for _, inst := range cfg.Instances {
		if inst.Role == system.RoleProjections {
			if len(inst.Engines) != 2 {
				t.Errorf("expected mixed projection pool of 2 engines, got %v", inst.Engines)
			}

			return
		}
	}

	t.Error("expected a projections instance")
}

func TestDeploymentShapeYAML_ManifestAcknowledgementsCache(t *testing.T) {
	t.Parallel()

	cfg, err := system.LoadConfig("testdata/deployment-manifest.yaml")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.ManifestPath == "" {
		t.Error("expected manifest_path to parse")
	}

	if len(cfg.AcknowledgeWarnings) != 1 || cfg.AcknowledgeWarnings[0] != "durability-downgrade:events" {
		t.Errorf("expected the acknowledged warning to parse, got %v", cfg.AcknowledgeWarnings)
	}

	for _, inst := range cfg.Instances {
		if inst.Role == system.RoleSourceOfTruth && inst.Cache != nil {
			if inst.Cache.Capacity != 10000 {
				t.Errorf("expected cache capacity 10000, got %d", inst.Cache.Capacity)
			}

			return
		}
	}

	t.Error("expected a cached source-of-truth instance")
}
