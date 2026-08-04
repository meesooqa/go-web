// Package lgr provides a universal logger constructor based on
// log/slog for use in different Go applications.
// Settings (level, format, output, rotation) are stored in YAML and
// loaded either independently via Load or as part of the application's
// main configuration (the Config structure is embedded into it as a field).
package lgr

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/meesooqa/go-cfg"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config defines the logger configuration. It is intended to be embedded
// into the application's main Config
// (for example, as the `Logger Config `yaml:"logger"` field)
// or loaded independently via Load.
type Config struct {
	// Level specifies the logging level: "debug", "info", "warn", or "error".
	Level string `yaml:"level" default:"info" env:"LOG_LEVEL"`

	// Format specifies the output format: "text" (human-readable) or "json"
	// (structured, for log aggregators).
	Format string `yaml:"format" default:"text" env:"LOG_FORMAT"`

	// Output specifies where logs are written: "stdout", "stderr", "file",
	//  or "both" (console and file simultaneously).
	Output string `yaml:"output" default:"stdout" env:"LOG_OUTPUT"`

	// FilePath is the path to the log file. Required if Output is "file"
	// or "both".
	FilePath string `yaml:"file_path" env:"LOG_FILE_PATH"`

	// AddSource includes the source file and line number in every log entry.
	AddSource bool `yaml:"add_source" default:"false" env:"LOG_ADD_SOURCE"`

	// Rotation defines log file rotation settings. It is used only when
	// Output is "file" or "both".
	Rotation RotationConfig `yaml:"rotation"`

	// Fields contains static fields attached to every log entry
	// (for example, service: "billing", env: "production"). This is useful
	// when multiple applications write to the same log aggregator.
	Fields map[string]string `yaml:"fields"`
}

// RotationConfig defines log file rotation settings (implemented using
// lumberjack). Zero values mean "not specified in YAML" - in that case,
// the values from the `default` tags are applied.
type RotationConfig struct {
	MaxSizeMB  int  `yaml:"max_size_mb" default:"100"`
	MaxBackups int  `yaml:"max_backups" default:"3"`
	MaxAgeDays int  `yaml:"max_age_days" default:"28"`
	Compress   bool `yaml:"compress" default:"true"`
}

// Load reads a YAML file (and optional .env files) using the cfg package
// and immediately constructs a ready-to-use logger. Use this function if
// your application does not have a single shared config.yml and logging
// configuration is the only configuration that needs to be loaded.
func Load(path string, envFiles ...string) (*slog.Logger, error) {
	c, err := cfg.Load[Config](path, envFiles...)
	if err != nil {
		return nil, err
	}
	return New(*c)
}

// New constructs an slog.Logger from a Config. Use this function
// if Config is part of your application's main configuration and has
// already been loaded separately (for example, via
// cfg.Load[AppConfig] with a Logger lgr.Config field).
func New(c Config) (*slog.Logger, error) {
	level, err := ParseLevel(c.Level)
	if err != nil {
		return nil, err
	}

	var handlers []slog.Handler

	switch c.Output {
	case "stdout", "":
		handlers = append(handlers, newHandler(os.Stdout, c.Format, level, c.AddSource))

	case "stderr":
		handlers = append(handlers, newHandler(os.Stderr, c.Format, level, c.AddSource))

	case "file":
		w, err := fileWriter(c.FilePath, c.Rotation)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, newHandler(w, c.Format, level, c.AddSource))

	case "both":
		w, err := fileWriter(c.FilePath, c.Rotation)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, newHandler(os.Stdout, c.Format, level, c.AddSource))
		handlers = append(handlers, newHandler(w, c.Format, level, c.AddSource))

	default:
		return nil, fmt.Errorf("lgr: unknown output value %q (valid values: stdout, stderr, file, both)", c.Output)
	}

	var handler slog.Handler
	if len(handlers) == 1 {
		handler = handlers[0]
	} else {
		handler = newMultiHandler(handlers...)
	}

	logger := slog.New(handler)

	if len(c.Fields) > 0 {
		attrs := make([]any, 0, len(c.Fields)*2)
		for k, v := range c.Fields {
			attrs = append(attrs, k, v)
		}
		logger = logger.With(attrs...)
	}

	return logger, nil
}

// newHandler creates a text or JSON slog handler that writes to w.
func newHandler(w io.Writer, format string, level slog.Level, addSource bool) slog.Handler {
	opts := &slog.HandlerOptions{Level: level, AddSource: addSource}
	if format == "json" {
		return slog.NewJSONHandler(w, opts)
	}
	return slog.NewTextHandler(w, opts)
}

// fileWriter returns an io.Writer backed by lumberjack - a log file with
// rotation based on size, number of backups, and file age.
func fileWriter(path string, r RotationConfig) (io.Writer, error) {
	if path == "" {
		return nil, fmt.Errorf("lgr: file_path must be set when output is file or both")
	}

	return &lumberjack.Logger{
		Filename:   path,
		MaxSize:    r.MaxSizeMB,
		MaxBackups: r.MaxBackups,
		MaxAge:     r.MaxAgeDays,
		Compress:   r.Compress,
	}, nil
}
