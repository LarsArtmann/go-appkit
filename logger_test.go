package appkit

import (
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestInitLogger_DefaultLevel(t *testing.T) {
	t.Parallel()

	logger, err := InitLogger(LoggerConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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

	logger, err := InitLogger(LoggerConfig{Level: LogLevelDebug})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !logger.Enabled(context.TODO(), slog.LevelDebug) {
		t.Error("expected debug level to be enabled")
	}
}

func TestInitLogger_InvalidLevel(t *testing.T) {
	t.Parallel()

	_, err := InitLogger(LoggerConfig{Level: LogLevel("banana")})
	if err == nil {
		t.Fatal("expected error for invalid level")
	}
}

func TestInitLogger_InvalidFormat(t *testing.T) {
	t.Parallel()

	_, err := InitLogger(LoggerConfig{Format: LogFormat("xml")})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestInitLogger_JSONFormat(t *testing.T) {
	t.Parallel()

	logger, err := InitLogger(LoggerConfig{Level: LogLevelInfo, Format: LogFormatJSON})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestInitLogger_TextFormat(t *testing.T) {
	t.Parallel()

	logger, err := InitLogger(LoggerConfig{Level: LogLevelInfo, Format: LogFormatText})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestInitLogger_AllLevels(t *testing.T) {
	t.Parallel()

	for _, level := range []LogLevel{LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError} {
		_, err := InitLogger(LoggerConfig{Level: level})
		if err != nil {
			t.Errorf("level %q: unexpected error: %v", level, err)
		}
	}
}

func TestLogLevel_SlogLevel_UnknownValue(t *testing.T) {
	t.Parallel()

	l := LogLevel("unknown")
	_, err := l.slogLevel()
	if err == nil {
		t.Fatal("expected error for unknown log level")
	}

	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should mention the invalid level, got: %v", err)
	}
}
