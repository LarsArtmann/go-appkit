package appkit

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// LoggerConfig controls how InitLogger builds the application logger.
type LoggerConfig struct {
	Level  string // "debug", "info", "warn", "error"
	Format string // "text", "json", "auto" (auto = text if output is a terminal, else json)
}

// InitLogger creates a structured slog.Logger using the supplied config.
// It panics if Level is not one of the supported values.
func InitLogger(cfg LoggerConfig) *slog.Logger {
	level := parseLevel(cfg.Level)
	w := os.Stderr
	var handler slog.Handler

	if cfg.Format == "json" || (cfg.Format == "auto" && !isTerminal(w)) {
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	}

	return slog.New(handler)
}

func parseLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "info", "":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		panic(fmt.Sprintf("unsupported log level %q", s))
	}
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}

	info, err := f.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}
