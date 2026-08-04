[English](README.md) | Русский

# cfg

Универсальный загрузчик YAML-конфигурации для Go-приложений с
поддержкой значений по умолчанию, переменных окружения (в том числе
из `.env`-файлов) и проверки обязательных полей.

Пакет не привязан к конкретному приложению — схему настроек (структуру)
вы определяете сами, в своём проекте. `cfg` отвечает только за механизм
загрузки.

## Установка

```bash
go get github.com/meesooqa/go-cfg
```

## Использование

```go
package main

import (
	"log"

	"github.com/meesooqa/go-cfg"
)

type Config struct {
	Server struct {
		Host string `yaml:"host" default:"0.0.0.0"`
		Port int    `yaml:"port" default:"8080" env:"SERVER_PORT"`
	} `yaml:"server"`

	Database struct {
		DSN string `yaml:"dsn" env:"DATABASE_DSN" required:"true"`
	} `yaml:"database"`
}

func main() {
	config, err := cfg.Load[Config]("config.yml", ".env")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s:%d", config.Server.Host, config.Server.Port)
}
```

## Теги структуры

| Тег        | Назначение                                             |
|------------|---------------------------------------------------------|
| `yaml`     | Имя поля в YAML-файле                                   |
| `default`  | Значение, если поле не задано в YAML                     |
| `env`      | Имя переменной окружения, переопределяющей значение      |
| `required` | `"true"` — ошибка при загрузке, если поле осталось пустым |

## Приоритет значений

```
переменная окружения (или .env) > YAML > default
```

## Разработка

```bash
go test ./...       # запустить тесты
go doc -all .       # посмотреть документацию
```

## Структура проекта

```
cfg.go          — точка входа, функция Load[T]
field.go        — конвертация string → тип поля через reflect
defaults.go     — применение тега default
env.go          — применение тега env и загрузка .env-файлов
required.go     — проверка тега required
*_test.go       — юнит-тесты
example_test.go — примеры использования (видны в godoc)
testdata/       — вспомогательные файлы для тестов
```