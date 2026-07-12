# go-i18n

A Go internationalization library with ICU MessageFormat support, deterministic fallbacks, and optional `net/http` locale detection

For development guidelines, see [AGENTS.md](AGENTS.md).
For internal contracts and design rules, start with [SPECS/00-overview.md](SPECS/00-overview.md).

## Features

- **ICU MessageFormat**: Use plural, select, and ordinal formatting through `github.com/kaptinlin/messageformat-go/mf1`.
- **Atomic loading**: Load maps, files, glob patterns, or embedded filesystems as immutable catalog generations safe for concurrent readers.
- **Deterministic fallbacks**: Resolve fallback chains at lookup time and use the configured default locale as the final fallback.
- **Lookup details**: Use `Lookup` to get rendered text, the resolved loaded template, matched locale, catalog locale, and result source.
- **Text and token keys**: Use token keys like `hello_world` or literal text keys with `GetX` context disambiguation.
- **HTTP integration**: Detect locales from query, cookie, header, or `Accept-Language`, then inject a request-scoped localizer.
- **Custom unmarshalers**: Keep JSON as the default, or plug in YAML, TOML, or INI parsing.

## Installation

Requires the Go version declared in `go.mod`.

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
	bundle, err := i18n.NewBundle(
		"en",
		i18n.WithLocales("zh-Hans"),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = bundle.LoadMessages(map[string]map[string]string{
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

- `NewBundle` requires the default locale, adds it to the supported set, and returns an error for invalid locale configuration.
- Do not repeat the default in `WithLocales` or an explicit fallback chain;
  `WithLocales` adds other supported locales and the bundle always tries its
  default after every configured fallback.
- `WithLocales`, `WithFallback`, `WithUnmarshaler`, `WithMessageFormatOptions`, `WithCustomFormatters`, and `WithStrictMode` configure optional bundle behavior.
- `WithMessageFormatOptions` accepts only the default/string return mode;
  `NewBundle` rejects values/parts mode because every rendering API returns a
  string.
- `NewLocalizer` picks the first supported locale match from its arguments, then falls back to the default locale.
- `SupportedLocales` and `IsLanguageSupported` expose the configured locale matcher state.

### Load translations

Use the loader that matches your source of truth:

- `LoadMessages` for in-memory maps
- `LoadFiles` for explicit files
- `LoadGlob` for file pattern expansion
- `LoadFS` for `go:embed` and other `fs.FS` implementations

Translation file names are normalized to locales, so names like `zh_Hans.json` and `zh-Hans.user.json` still resolve to `zh-Hans`.
Configured locale mistakes fail loudly: invalid construction locales make `NewBundle` return an error, and loading an unconfigured map locale or file locale returns an error without changing the existing catalog.
Within one `LoadMessages` or file-loader call, locale aliases or fragments may
declare disjoint keys; repeating a key for the same canonical locale makes the
whole batch fail. File-loader errors identify both source paths. A later
successful loader call may intentionally replace an existing key. File
MessageFormat errors identify the source path, canonical locale, and key.

### Render translations

Use `Get` for the normal translation path, `GetX` for context-disambiguated text keys, and `Format` for dynamic messages that are not stored in translation files.

```go
localizer := bundle.NewLocalizer("zh-Hans")

fmt.Println(localizer.Get("hello", i18n.Vars{"name": "Lin"}))
fmt.Println(localizer.GetX("Post", "verb"))
```

### Inspect fallback behavior

Use `Lookup` when you need to know where a translation came from and which loaded template supplied it.

```go
result, err := localizer.Lookup("hello", i18n.Vars{"name": "Lin"})
if err != nil {
	log.Printf("render translation: %v", err)
}
fmt.Println(result.Text)
fmt.Println(result.Template)
fmt.Println(result.MatchedLocale)
fmt.Println(result.CatalogLocale)
fmt.Println(result.Source)
```

`Template` is the resolved loaded template: the raw MessageFormat text from the direct or fallback catalog entry that supplied `Text`. It is empty when `Source` is `missing`. Use `GetTemplate` when you only need that resolved loaded template without formatting.

`MatchedLocale` is the locale selected for the localizer. `CatalogLocale` is the loaded catalog locale that supplied the text; it is empty when `Source` is `missing`. `TranslationSource` reports one of `direct`, `fallback`, or `missing`.

If a loaded template cannot be formatted, `Lookup` returns its raw text and
provenance together with a wrapped error. A missing translation is not an error.
Use `Get` when raw fallback without error handling is the desired behavior.

Use `Has` and `Keys` for direct locale contents only. They do not include inherited fallback keys.

### Detect request locales

Use `NewDetector` to resolve the best locale from HTTP request inputs.

```go
detector, err := i18n.NewDetector(
	bundle,
	i18n.WithDetectorPriority(
		i18n.DetectorSourceQuery,
		i18n.DetectorSourceCookie,
		i18n.DetectorSourceHeader,
		i18n.DetectorSourceAccept,
	),
)
if err != nil {
	return err
}

locale := detector.DetectLocale(r)
localizer := bundle.NewLocalizer(locale)
```

Use `WithDetectorQueryParam`, `WithDetectorCookieName`, and `WithDetectorHeaderName` to customize source names.
`NewDetector` rejects a nil bundle, a nil option, or an unknown priority source
at setup; empty priority input keeps the default source order.
Use `MatchAvailableLocale` when you only need `Accept-Language` matching. Pass
all header field values so their quality weights are evaluated together:

```go
locale := bundle.MatchAvailableLocale(r.Header.Values("Accept-Language")...)
```

### Inject a localizer into `net/http`

Use the optional middleware package to attach a request-scoped localizer to the request context.

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kaptinlin/go-i18n"
	"github.com/kaptinlin/go-i18n/middleware"
)

func main() {
	bundle, err := i18n.NewBundle("en", i18n.WithLocales("zh-Hans"))
	if err != nil {
		log.Fatal(err)
	}
	if err := bundle.LoadMessages(map[string]map[string]string{
		"en": {"hello": "Hello"},
	}); err != nil {
		log.Fatal(err)
	}

	localize, err := middleware.HTTPMiddleware(
		bundle,
		i18n.WithDetectorPriority(i18n.DetectorSourceAccept),
	)
	if err != nil {
		log.Fatal(err)
	}
	handler := localize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		localizer, ok := i18n.LocalizerFromContext(r.Context())
		if !ok {
			http.Error(w, "missing localizer", http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, localizer.Get("hello"))
	}))

	log.Fatal(http.ListenAndServe(":8080", handler))
}
```

### Use custom unmarshalers

JSON is the default. Override it when your translation files use another format.

```go
import (
	"log"

	"github.com/goccy/go-yaml"
)

bundle, err := i18n.NewBundle(
	"en",
	i18n.WithLocales("zh-Hans"),
	i18n.WithUnmarshaler(yaml.Unmarshal),
)
if err != nil {
	log.Fatal(err)
}
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

task verify
```

## Contributing

- Keep examples runnable and user-facing.
- Run `task fmt`, `task vet`, `task lint`, `task test` before shipping changes.
- Keep `README.md`, `CLAUDE.md`, and [SPECS/00-overview.md](SPECS/00-overview.md) aligned when public behavior changes.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
