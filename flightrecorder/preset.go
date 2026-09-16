package flightrecorder

import (
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
