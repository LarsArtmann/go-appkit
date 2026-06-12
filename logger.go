package appkit

import (
	"errors"
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

var (
	errUnsupportedLogLevel  = errors.New("unsupported log level")
	errUnsupportedLogFormat = errors.New("unsupported log format")
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
		return 0, fmt.Errorf("%w: %q", errUnsupportedLogLevel, l)
	}
}

func (f LogFormat) isJSON(writer io.Writer) (bool, error) {
	switch f {
	case LogFormatJSON:
		return true, nil
	case LogFormatText, "":
		return false, nil
	case LogFormatAuto:
		return !isTerminal(writer), nil
	default:
		return false, fmt.Errorf("%w: %q", errUnsupportedLogFormat, f)
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

	writer := os.Stderr

	useJSON, err := cfg.Format.isJSON(writer)
	if err != nil {
		return nil, err
	}

	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:       level,
		AddSource:   false,
		ReplaceAttr: nil,
	}

	if useJSON {
		handler = slog.NewJSONHandler(writer, opts)
	} else {
		handler = slog.NewTextHandler(writer, opts)
	}

	return slog.New(handler), nil
}

func isTerminal(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}
