package appkit

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
	LogFormatAuto LogFormat = "auto"
)

func (l LogLevel) slogLevel() (slog.Level, error) {
	switch l {
	case LogLevelDebug:
		return slog.LevelDebug, nil
	case LogLevelInfo, "":
		return slog.LevelInfo, nil
	case LogLevelWarn:
		return slog.LevelWarn, nil
	case LogLevelError:
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unsupported log level %q", l)
	}
}

func (f LogFormat) isJSON(w io.Writer) (bool, error) {
	switch f {
	case LogFormatJSON:
		return true, nil
	case LogFormatText, "":
		return false, nil
	case LogFormatAuto:
		return !isTerminal(w), nil
	default:
		return false, fmt.Errorf("unsupported log format %q", f)
	}
}

type LoggerConfig struct {
	Level  LogLevel
	Format LogFormat
}

func InitLogger(cfg LoggerConfig) (*slog.Logger, error) {
	level, err := cfg.Level.slogLevel()
	if err != nil {
		return nil, err
	}

	w := os.Stderr

	useJSON, err := cfg.Format.isJSON(w)
	if err != nil {
		return nil, err
	}

	var handler slog.Handler

	opts := &slog.HandlerOptions{Level: level}

	if useJSON {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	return slog.New(handler), nil
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
