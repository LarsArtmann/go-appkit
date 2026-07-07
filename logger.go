package appkit

import (
	"log/slog"
	"os"

	"github.com/charmbracelet/log"
	errorfamily "github.com/larsartmann/go-error-family"
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
		return 0, errorfamily.Newf(errorfamily.Rejection, "log.level_invalid", "unsupported log level: %q", l)
	}
}

func (f LogFormat) formatter() (log.Formatter, error) {
	switch f {
	case LogFormatJSON:
		return log.JSONFormatter, nil
	case LogFormatText, LogFormatAuto, "":
		return log.TextFormatter, nil
	default:
		return 0, errorfamily.Newf(errorfamily.Rejection, "log.format_invalid", "unsupported log format: %q", f)
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

	formatter, err := cfg.Format.formatter()
	if err != nil {
		return nil, err
	}

	cl := log.NewWithOptions(os.Stderr, log.Options{
		Level:     log.Level(level),
		Formatter: formatter,
	})

	return slog.New(cl), nil
}
