package lgr

import (
	"log/slog"
	"os"
)

// Option configures a quick logger created via NewConsole
// or NewFile for cases where a full Config is excessive
// (scripts, CLI utilities, tests).
type Option func(*quickOptions)

type quickOptions struct {
	format    string
	addSource bool
	rotation  RotationConfig
}

func defaultQuickOptions() quickOptions {
	return quickOptions{
		format: "text",
		rotation: RotationConfig{
			MaxSizeMB:  100,
			MaxBackups: 3,
			MaxAgeDays: 28,
			Compress:   true,
		},
	}
}

// WithJSON switches the output format to JSON (by default, text is used
// for NewConsole and JSON for NewFile).
func WithJSON() Option {
	return func(o *quickOptions) { o.format = "json" }
}

// WithText switches the output format to text.
func WithText() Option {
	return func(o *quickOptions) { o.format = "text" }
}

// WithSource adds the file and line number of the call site to each record.
func WithSource() Option {
	return func(o *quickOptions) { o.addSource = true }
}

// WithRotation sets rotation parameters for NewFile (by default:
// 100 MB, 3 backups, 28 days, compression enabled).
func WithRotation(r RotationConfig) Option {
	return func(o *quickOptions) { o.rotation = r }
}

// NewConsole creates a logger that writes to stdout without reading
// configuration from a file. Suitable for CLI utilities, scripts,
// and quick startup.
func NewConsole(level slog.Level, opts ...Option) *slog.Logger {
	o := defaultQuickOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return slog.New(newHandler(os.Stdout, o.format, level, o.addSource))
}

// NewFile creates a logger that writes to the file at path, with
// default rotation settings. It uses JSON format by default (which is
// usually more convenient for files); this can be overridden with WithText.
func NewFile(path string, level slog.Level, opts ...Option) (*slog.Logger, error) {
	o := defaultQuickOptions()
	o.format = "json"
	for _, opt := range opts {
		opt(&o)
	}

	w, err := fileWriter(path, o.rotation)
	if err != nil {
		return nil, err
	}

	return slog.New(newHandler(w, o.format, level, o.addSource)), nil
}
