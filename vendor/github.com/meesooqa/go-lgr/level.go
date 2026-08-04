package lgr

import (
	"fmt"
	"log/slog"
	"strings"
)

// ParseLevel converts a string logging level name
// ("debug", "info", "warn"/"warning", "error") into a slog.Level.
// The case is ignored. An empty string is treated as "info".
// The logging level is stored in YAML as a string
// because slog.Level itself does not have
// a direct textual representation in YAML.
func ParseLevel(raw string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("lgr: unknown logging level %q (valid values: debug, info, warn, error)", raw)
	}
}
