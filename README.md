# go-i18n

A Go internationalization library with ICU MessageFormat support, deterministic fallbacks, and optional `net/http` locale detection

For development guidelines, see [AGENTS.md](AGENTS.md).
For internal contracts and design rules, start with [SPECS/00-overview.md](SPECS/00-overview.md).

## Features

- **ICU MessageFormat**: Use plural, select, and ordinal formatting through `github.com/kaptinlin/messageformat-go/v1`.
- **Flexible loading**: Load translations from maps, files, glob patterns, or embedded filesystems.
- **Deterministic fallbacks**: Root fallback chains in the configured default locale.
- **Lookup details**: Use `Lookup` to get the rendered text, resolved locale, and result source.
- **Text and token keys**: Use token keys like `hello_world` or literal text keys with `GetX` context disambiguation.
- **HTTP integration**: Detect locales from query, cookie, header, or `Accept-Language`, then inject a request-scoped localizer.
- **Custom unmarshalers**: Keep JSON as the default, or plug in YAML, TOML, or INI parsing.

## Installation

Requires Go 1.26+.

```bash
go get github.com/kaptinlin/go-i18n@latest
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/go-i18n"
)

func main() {
	bundle := i18n.NewBundle(
		i18n.WithDefaultLocale("en"),
		i18n.WithLocales("en", "zh-Hans"),
	)

	err := bundle.LoadMessages(map[string]map[string]string{
		"en": {"hello": "Hello, {name}!"},
		"zh-Hans": {"hello": "你好，{name}！"},
	})
	if err != nil {
		log.Fatal(err)
	}

	localizer := bundle.NewLocalizer("zh-CN")
	fmt.Println(localizer.Get("hello", i18n.Vars{"name": "World"}))
}
```

## Core API

### Bundle and localizer

- `NewBundle` constructs the shared translation bundle.
- `WithDefaultLocale`, `WithLocales`, `WithFallback`, `WithUnmarshaler`, `WithMessageFormatOptions`, `WithCustomFormatters`, and `WithStrictMode` configure bundle behavior.
- `NewLocalizer` picks the first matching locale from its arguments, then falls back to the default locale.
- `SupportedLocales` and `IsLanguageSupported` expose the configured locale matcher state.

### Load translations

Use the loader that matches your source of truth:

- `LoadMessages` for in-memory maps
- `LoadFiles` for explicit files
- `LoadGlob` for file pattern expansion
- `LoadFS` for `go:embed` and other `fs.FS` implementations

Translation file names are normalized to locales, so names like `zh_Hans.json` and `zh-Hans.user.json` still resolve to `zh-Hans`.

### Render translations

Use `Get` for the normal translation path, `GetX` for context-disambiguated text keys, and `Format` for dynamic messages that are not stored in translation files.

```go
localizer := bundle.NewLocalizer("zh-Hans")

fmt.Println(localizer.Get("hello", i18n.Vars{"name": "Lin"}))
fmt.Println(localizer.GetX("Post", "verb"))
```

### Inspect fallback behavior

Use `Lookup` when you need to know where a translation came from.

```go
result := localizer.Lookup("hello", i18n.Vars{"name": "Lin"})
fmt.Println(result.Text)
fmt.Println(result.Locale)
fmt.Println(result.Source)
```

`TranslationSource` reports one of `direct`, `fallback`, or `missing`.

Use `Has` and `Keys` for direct locale contents only. They do not include inherited fallback keys.

### Detect request locales

Use `NewDetector` to resolve the best locale from HTTP request inputs.

```go
detector := i18n.NewDetector(
	bundle,
	i18n.WithDetectorPriority(
		i18n.DetectorSourceQuery,
		i18n.DetectorSourceCookie,
		i18n.DetectorSourceHeader,
		i18n.DetectorSourceAccept,
	),
)

locale := detector.DetectLocale(r)
localizer := bundle.NewLocalizer(locale)
```

Use `WithDetectorQueryParam`, `WithDetectorCookieName`, and `WithDetectorHeaderName` to customize source names.
Use `MatchAvailableLocale` when you only need `Accept-Language` matching.

### Inject a localizer into `net/http`

Use the optional middleware package to attach a request-scoped localizer to the request context.

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/kaptinlin/go-i18n"
	"github.com/kaptinlin/go-i18n/middleware"
)

func main() {
	bundle := i18n.NewBundle(i18n.WithDefaultLocale("en"), i18n.WithLocales("en", "zh-Hans"))
	_ = bundle.LoadMessages(map[string]map[string]string{
		"en": {"hello": "Hello"},
	})

	handler := middleware.HTTPMiddleware(bundle)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		localizer, ok := i18n.LocalizerFromContext(r.Context())
		if !ok {
			http.Error(w, "missing localizer", http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, localizer.Get("hello"))
	}))

	http.ListenAndServe(":8080", handler)
}
```

### Use custom unmarshalers

JSON is the default. Override it when your translation files use another format.

```go
import "github.com/goccy/go-yaml"

bundle := i18n.NewBundle(
	i18n.WithDefaultLocale("en"),
	i18n.WithLocales("en", "zh-Hans"),
	i18n.WithUnmarshaler(yaml.Unmarshal),
)
```

See `examples/yml`, `examples/toml`, and `examples/ini` for complete loaders.

## Examples

Run the package examples directly:

```bash
go run ./examples/basic
go run ./examples/files
go run ./examples/embed
go run ./examples/glob
go run ./examples/icu
go run ./examples/source_locale
go run ./examples/text
```

Additional examples cover nested files and custom unmarshalers under `examples/`.

## API Reference

See [pkg.go.dev/github.com/kaptinlin/go-i18n](https://pkg.go.dev/github.com/kaptinlin/go-i18n) for the exported API.

## Development

Run the project commands from the repository root:

```bash
task test
task test-coverage
task bench
task fmt
task vet
task lint
task markdownlint
task verify
```

## Contributing

- Keep examples runnable and user-facing.
- Run `task fmt`, `task vet`, `task lint`, `task test`, and `task markdownlint` before shipping changes.
- Keep `README.md`, `CLAUDE.md`, and [SPECS/00-overview.md](SPECS/00-overview.md) aligned when public behavior changes.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
