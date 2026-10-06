package flightrecorder

import (
	"fmt"
	"log/slog"
	"time"

	fr "github.com/larsartmann/go-flightrecorder"
)

const (
	// presetCompression is the gzip level for snapshot files (3 = a middle
	// ground between CPU cost and trace-file size).
	presetCompression = 3

	// presetMinAge discards traces shorter than this: sub-10s traces are
	// rarely diagnosable and burn the retention budget.
	presetMinAge = 10 * time.Second
)

// OpsRecorderPreset returns the documented production fr options: snapshots
// land in dir, capped at maxSnapshots files and maxBytes total, gzip
// compression level 3, with a 10s minimum trace age (shorter traces are
// rarely diagnosable and burn the retention budget).
//
// This is a CONVENIENCE PRESET, not magic: the returned options are plain
// go-flightrecorder options — append yours (e.g. fr.WithMetrics(hook),
// fr.WithLogger(hook)) after spreading. Example:
//
//	rec, err := fr.New(append(flightrecorder.OpsRecorderPreset(
//	    "/var/lib/appkit/traces", 5, 64<<20,
//	), myMetricsHook)...)
func OpsRecorderPreset(dir string, maxSnapshots int, maxBytes uint64) []fr.Option {
	return []fr.Option{
		fr.WithSnapshotDir(dir),
		fr.WithSnapshotPrefix("trace"),
		fr.WithMaxSnapshots(maxSnapshots),
		fr.WithMaxBytes(maxBytes),
		fr.WithCompression(presetCompression),
		fr.WithMinAge(presetMinAge),
	}
}

// OpsRecorderLoggerPreset is [OpsRecorderPreset] plus [fr.WithLogger]: the
// recorder's lifecycle events and — critically — RETENTION FAILURES land in
// the service log instead of vanishing silently (a full disk or a permission
// error would otherwise eat new snapshots with no signal). Pass the service
// logger; a nil logger skips the hook (the preset stays usable, but you give
// up the failure visibility this variant exists for).
//
// The returned slice carries 7 options when a logger is set (6 otherwise);
// append yours after spreading, same as the base preset:
//
//	rec, err := fr.New(flightrecorder.OpsRecorderLoggerPreset(
//	    "/var/lib/appkit/traces", 5, 64<<20, svc.Logger,
//	)...)
func OpsRecorderLoggerPreset(dir string, maxSnapshots int, maxBytes uint64, log *slog.Logger) []fr.Option {
	opts := OpsRecorderPreset(dir, maxSnapshots, maxBytes)

	if log != nil {
		opts = append(opts, fr.WithLogger(func(format string, args ...any) {
			log.Info(fmt.Sprintf(format, args...))
		}))
	}

	return opts
}
