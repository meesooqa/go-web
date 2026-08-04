English | [Русский](README_ru.md)

# lgr

A universal logger builder based on `log/slog` for use across different Go
applications. Configuration is stored in YAML and can be loaded either
independently (`Load`) or as part of the application's main configuration
using the [`cfg`](https://github.com/meesooqa/go-cfg) package (by embedding the
`Config` struct as one of the fields).

## Installation

```bash
go get github.com/meesooqa/go-lgr
```

## Configuration (`Config`)

| Field          | YAML Key      | Default      | Description                                                  |
|----------------|---------------|--------------|--------------------------------------------------------------|
| `Level`        | `level`       | `info`       | `debug` \| `info` \| `warn` \| `error`                       |
| `Format`       | `format`      | `text`       | `text` (human-readable) \| `json` (structured)               |
| `Output`       | `output`      | `stdout`     | `stdout` \| `stderr` \| `file` \| `both`                     |
| `FilePath`     | `file_path`   | —            | Path to the log file. Required for `output: file` / `both`.  |
| `AddSource`    | `add_source`  | `false`      | Include the caller's `file:line` in every log entry.         |
| `Rotation.*`   | `rotation`    | see below    | Log file rotation settings (via lumberjack).                 |
| `Fields`       | `fields`      | —            | Static fields added to every log entry.                      |

Default rotation (`rotation`) settings: `max_size_mb: 100`,
`max_backups: 3`, `max_age_days: 28`, `compress: true`.

The logging level is stored in YAML as a string (`"debug"`, `"info"`, etc.)
because `slog.Level` is a numeric type without a native textual YAML
representation. Conversion is handled by `lgr.ParseLevel`.

## Usage

### 1. Quick console logger (without configuration)

```go
logger := lgr.NewConsole(slog.LevelDebug)
logger.Info("service started")
```

### 2. File logger with default rotation

```go
logger, err := lgr.NewFile("/var/log/app.log", slog.LevelInfo)
if err != nil {
    log.Fatal(err)
}
logger.Info("started", "version", "1.2.3")
```

### 3. Build a logger from `Config`

```go
logger, err := lgr.New(lgr.Config{
    Level:    "info",
    Format:   "json",
    Output:   "both",
    FilePath: "/var/log/app.log",
})
```

### 4. Load configuration from a dedicated YAML file

```go
logger, err := lgr.Load("logging.yml", ".env")
```

### 5. Embed into the application's main configuration (via `cfg`)

This is the preferred approach if your application already uses the `cfg`
package for its configuration:

```go
type AppConfig struct {
    Server ServerConfig `yaml:"server"`
    Logger lgr.Config   `yaml:"logger"`
}

func main() {
    appConfig, err := cfg.Load[AppConfig]("config.yml")
    if err != nil {
        log.Fatal(err)
    }

    logger, err := lgr.New(appConfig.Logger)
    if err != nil {
        log.Fatal(err)
    }

    slog.SetDefault(logger) // slog.Info(...) now works globally
}
```

```yaml
# config.yml
server:
  host: "0.0.0.0"
  port: 8080

logger:
  level: info
  format: json
  output: both
  file_path: /var/log/myapp/app.log
  fields:
    service: "myapp"
    env: "production"
```

### 6. Request-scoped logger

```go
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestLogger := logger.With("request_id", uuid.NewString())
        ctx := lgr.NewContext(r.Context(), requestLogger)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func handler(w http.ResponseWriter, r *http.Request) {
    lgr.FromContext(r.Context()).Info("handling request")
}
```

### 7. Child loggers for individual modules

This is a built-in `slog` feature and does not require any additional support
from `go-lgr`:

```go
dbLogger := logger.With("component", "db")
dbLogger.Debug("query executed", "duration_ms", 42)
```

## Configuration source precedence (via `cfg`)

When `Config` is embedded into the application's main configuration, the same
precedence rules as in `cfg` apply: `env` > `yaml` > `default`.

For example, `LOG_LEVEL=debug` overrides `level: info` from the YAML file.

## Development

```bash
go test ./...
```

## Project structure

```
lgr.go              — Config, New, Load, fileWriter
level.go            — ParseLevel (string → slog.Level)
multi.go            — multiHandler (duplicates log records to multiple handlers)
console_file.go     — NewConsole, NewFile, functional options
context.go          — NewContext, FromContext
*_test.go           — Unit and integration tests
example_test.go     — Examples for pkg.go.dev and GoDoc
testdata/           — Auxiliary files for tests and documentation
```