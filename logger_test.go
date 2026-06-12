package appkit

import (
	"context"
	"log/slog"
	"testing"
)

func TestInitLogger_DefaultLevel(t *testing.T) {
	t.Parallel()

	logger := InitLogger(LoggerConfig{})
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	ctx := context.TODO()

	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("expected info level to be enabled by default")
	}

	if logger.Enabled(ctx, slog.LevelDebug) {
		t.Error("expected debug level to be disabled by default")
	}
}

func TestInitLogger_DebugLevel(t *testing.T) {
	t.Parallel()

	logger := InitLogger(LoggerConfig{Level: "debug"})
	if !logger.Enabled(context.TODO(), slog.LevelDebug) {
		t.Error("expected debug level to be enabled")
	}
}

func TestInitLogger_InvalidLevelPanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid level")
		}
	}()

	InitLogger(LoggerConfig{Level: "banana"})
}
