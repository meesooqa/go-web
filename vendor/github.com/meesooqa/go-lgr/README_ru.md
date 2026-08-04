[English](README.md) | Русский

# lgr

Универсальный конструктор логгеров на основе `log/slog` для
использования в разных Go-приложениях. Настройки хранятся в YAML и
загружаются либо самостоятельно (`Load`), либо как часть общего
конфига приложения через пакет [`cfg`](https://github.com/meesooqa/go-cfg) (структура `Config`
встраивается как одно из полей).

## Установка

```bash
go get github.com/meesooqa/go-lgr
```

## Настройки (Config)

| Поле         | YAML-ключ     | По умолчанию | Описание                                              |
|--------------|---------------|--------------|--------------------------------------------------------|
| `Level`      | `level`       | `info`       | `debug` \| `info` \| `warn` \| `error`                   |
| `Format`     | `format`      | `text`       | `text` (человекочитаемый) \| `json` (структурированный) |
| `Output`     | `output`      | `stdout`     | `stdout` \| `stderr` \| `file` \| `both`                  |
| `FilePath`   | `file_path`   | —            | Путь к файлу. Обязателен для `output: file` / `both`     |
| `AddSource`  | `add_source`  | `false`      | Добавлять `file:line` вызова в каждую запись             |
| `Rotation.*` | `rotation`    | см. ниже     | Ротация файла лога (через lumberjack)                     |
| `Fields`     | `fields`      | —            | Статичные поля, добавляемые в каждую запись               |

Ротация (`rotation`) по умолчанию: `max_size_mb: 100`, `max_backups: 3`,
`max_age_days: 28`, `compress: true`.

Уровень логирования хранится в YAML как строка (`"debug"`, `"info"` и
т.д.), поскольку `slog.Level` — это числовой тип без прямого текстового
представления в YAML. Конвертация — через `lgr.ParseLevel`.

## Способы использования

### 1. Быстрый консольный логгер (без конфига)

```go
logger := lgr.NewConsole(slog.LevelDebug)
logger.Info("service started")
```

### 2. Логгер в файл с ротацией по умолчанию

```go
logger, err := lgr.NewFile("/var/log/app.log", slog.LevelInfo)
if err != nil {
    log.Fatal(err)
}
logger.Info("started", "version", "1.2.3")
```

### 3. Полная сборка из Config

```go
logger, err := lgr.New(lgr.Config{
    Level:  "info",
    Format: "json",
    Output: "both",
    FilePath: "/var/log/app.log",
})
```

### 4. Самостоятельная загрузка из своего YAML-файла

```go
logger, err := lgr.Load("logging.yml", ".env")
```

### 5. Встраивание в общий конфиг приложения (через cfg)

Самый предпочтительный способ, если в приложении уже используется
пакет `cfg` для остальных настроек:

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

    slog.SetDefault(logger) // теперь slog.Info(...) работает глобально
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

### 6. Логгер в контексте запроса

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

### 7. Дочерние логгеры для модулей

Встроенная возможность `slog`, не требует ничего от `go-lgr`:

```go
dbLogger := logger.With("component", "db")
dbLogger.Debug("query executed", "duration_ms", 42)
```

## Приоритет источников (через cfg)

Если Config встроена в общий конфиг приложения, действуют те же
правила, что и в `cfg`: `env` > `yaml` > `default`. Например,
`LOG_LEVEL=debug` переопределит `level: info` из файла.

## Разработка

```bash
go test ./...
```

## Структура файлов

```
lgr.go              — Config, New, Load, fileWriter
level.go            — ParseLevel (string → slog.Level)
multi.go            — multiHandler (дублирование записи в несколько обработчиков)
console_file.go     — NewConsole, NewFile, functional options
context.go          — NewContext, FromContext
*_test.go           — юнит- и интеграционные тесты
example_test.go      — примеры для godoc
testdata/            — вспомогательные файлы для тестов и документации
```
