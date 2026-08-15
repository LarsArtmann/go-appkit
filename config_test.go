package appkit

import (
	"testing"
	"time"
)

func TestValidate_NegativeTimeouts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  ServiceConfig
	}{
		{"negative ReadTimeout", ServiceConfig{Addr: ":0", ReadTimeout: -2 * time.Millisecond}},
		{"negative ReadHeaderTimeout", ServiceConfig{Addr: ":0", ReadHeaderTimeout: -2 * time.Millisecond}},
		{"negative WriteTimeout", ServiceConfig{Addr: ":0", WriteTimeout: -2 * time.Millisecond}},
		{"negative IdleTimeout", ServiceConfig{Addr: ":0", IdleTimeout: -1}},
		{"NoTimeout on ReadHeaderTimeout", ServiceConfig{Addr: ":0", ReadHeaderTimeout: NoTimeout}},
		{"NoTimeout on IdleTimeout", ServiceConfig{Addr: ":0", IdleTimeout: NoTimeout}},
		{"zero ShutdownTimeout", ServiceConfig{Addr: ":0", ShutdownTimeout: -1}},
		{"negative DrainDelay", ServiceConfig{Addr: ":0", DrainDelay: -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewService(tt.cfg)
			if err == nil {
				t.Fatalf("expected validation error for %s", tt.name)
			}
		})
	}
}

func TestValidate_ValidConfigs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  ServiceConfig
	}{
		{"all defaults", DefaultServiceConfig()},
		{"custom addr", ServiceConfig{Addr: "localhost:9999"}},
		{"custom timeouts", ServiceConfig{
			Addr:              ":0",
			ReadTimeout:       5 * time.Second,
			ReadHeaderTimeout: 3 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       30 * time.Second,
		}},
		{"zero DrainDelay", ServiceConfig{Addr: ":0", DrainDelay: 0}},
		{"debug logging", ServiceConfig{Addr: ":0", LogLevel: LogLevelDebug}},
		{"NoTimeout on Read and Write (SSE)", ServiceConfig{
			Addr:         ":0",
			ReadTimeout:  NoTimeout,
			WriteTimeout: NoTimeout,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc, err := NewService(tt.cfg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			_ = svc.Close()
		})
	}
}

func TestApplyDefaults_AllFields(t *testing.T) {
	t.Parallel()

	cfg := ServiceConfig{}
	cfg.applyDefaults()

	if cfg.Addr != defaultAddr {
		t.Errorf("Addr = %q, want %q", cfg.Addr, defaultAddr)
	}

	if cfg.LogLevel != LogLevelInfo {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, LogLevelInfo)
	}

	if cfg.LogFormat != LogFormatAuto {
		t.Errorf("LogFormat = %q, want %q", cfg.LogFormat, LogFormatAuto)
	}

	if cfg.ReadTimeout != defaultReadTimeout {
		t.Errorf("ReadTimeout = %v, want %v", cfg.ReadTimeout, defaultReadTimeout)
	}

	if cfg.WriteTimeout != defaultWriteTimeout {
		t.Errorf("WriteTimeout = %v, want %v", cfg.WriteTimeout, defaultWriteTimeout)
	}

	if cfg.IdleTimeout != defaultIdleTimeout {
		t.Errorf("IdleTimeout = %v, want %v", cfg.IdleTimeout, defaultIdleTimeout)
	}

	if cfg.ShutdownTimeout != defaultShutdownTimeout {
		t.Errorf("ShutdownTimeout = %v, want %v", cfg.ShutdownTimeout, defaultShutdownTimeout)
	}

	if cfg.DrainDelay != defaultDrainDelay {
		t.Errorf("DrainDelay = %v, want %v", cfg.DrainDelay, defaultDrainDelay)
	}
}
