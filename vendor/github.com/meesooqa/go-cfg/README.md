English | [Русский](README_ru.md)

# cfg

A universal YAML configuration loader for Go applications with support for
default values, environment variables (including values from `.env` files),
and validation of required fields.

The package is not tied to any specific application — you define the
configuration schema (structure) yourself in your own project. `cfg` is
responsible only for the loading mechanism.

## Installation

```bash
go get github.com/meesooqa/go-cfg
```

## Usage

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

## Struct Tags

| Тег        | Назначение                                           |
|------------|------------------------------------------------------|
| `yaml`     | Field name in the YAML file                          |
| `default`  | Value used if the field is not specified in YAML     |
| `env`      | Environment variable name overriding the value       |
| `required` | `"true"` — loading fails if the field remains empty  |

## Value Priority

```
environment variable (or .env) > YAML > default
```

## Development

```bash
go test ./...       # run tests
go doc -all .       # view documentation
```

## Project Structure

```
cfg.go          — entry point, Load[T] function
field.go        — string → field type conversion using reflect
defaults.go     — applying the default tag
env.go          — applying the env tag and loading .env files
required.go     — validating the required tag
*_test.go       — unit tests
example_test.go — usage examples (visible in godoc)
testdata/       — auxiliary test files
```