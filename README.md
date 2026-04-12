# i18n (Go)

`kaptinlin/go-i18n` is a high-performance, modern localization and internationalization library for Go.

## Features

- **Token-based** (`hello_world`) and **Text-based** (`Hello, world!`) translation
- **High Performance**: Optimized with Go 1.26 features (slices, maps, built-in functions)
- **ICU MessageFormat v1**: Full support with [kaptinlin/messageformat-go](https://github.com/kaptinlin/messageformat-go)
- **Flexible Loading**: From maps, files, glob patterns, or `go:embed`
- **Deterministic Fallbacks**: Recursive fallback chains rooted in the default locale
- **Custom Formatters**: Extensible formatting system for complex use cases
- **Accept-Language**: Built-in HTTP header parsing support

## Index

-   [Installation](#installation)
-   [Getting Started](#getting-started)
-   [Advanced Configuration](#advanced-configuration)
    -   [Custom Formatters](#custom-formatters)
    -   [Strict Mode](#strict-mode)
    -   [MessageFormat Options](#messageformat-options)
-   [Loading Methods](#loading-methods)
    -   [Load from Go map](#load-from-go-map)
    -   [Load from Files](#load-from-files)
    -   [Load from Glob Matching Files](#load-from-glob-matching-files)
    -   [Load from Embedded Files](#load-from-embedded-files)
-   [Translations](#translations)
    -   [Passing Data to Translation](#passing-data-to-translation)
    -   [Dynamic Formatting](#dynamic-formatting)
-   [Translation Lookup](#translation-lookup)
    -   [Detecting Fallback vs Direct Hit](#detecting-fallback-vs-direct-hit)
    -   [Google AIP-193 Compliance](#google-aip-193-compliance)
    -   [Translation Coverage Analytics](#translation-coverage-analytics)
-   [Pluralization](#pluralization)
-   [Text-based Translations](#text-based-translations)
    -   [Disambiguation by context](#disambiguation-by-context)
    -   [Act as fallback](#act-as-fallback)
-   [Fallbacks](#fallbacks)
-   [Introspection](#introspection)
-   [Custom Unmarshaler](#custom-unmarshaler)
    -   [YAML Unmarshaler](#yaml-unmarshaler)
    -   [TOML Unmarshaler](#toml-unmarshaler)
    -   [INI Unmarshaler](#ini-unmarshaler)
-   [Parse Accept-Language](#parse-accept-language)
-   [Performance](#performance)

&nbsp;

## Installation

```bash
go get github.com/kaptinlin/go-i18n@latest
```

&nbsp;

## Getting started

Create a folder named `./locales` and put some `YAML`, `TOML`, `INI` or `JSON` files.

```sh
│   main.go
└───locales
    ├───en.json
    └───zh-Hans.json
```

Now, put the key-values content for each locale, e.g.

File: `locales/en.json`

```json
{
  "hello": "Hello, {name}"
}
```

File: `locales/zh-Hans.json`

```json
{
  "hello": "你好, {name}"
}
```

File: `main.go`

```go
package main

import (
    "github.com/kaptinlin/go-i18n"
    "fmt"
)

func main() {
    bundle := i18n.NewBundle(
        i18n.WithDefaultLocale("en"),
        i18n.WithLocales("en", "zh-Hans"),
    )

    err := bundle.LoadFiles("./locales/zh-Hans.json", "./locales/en.json")
    if err != nil {
        fmt.Println(err)
    }

    localizer := bundle.NewLocalizer("zh-Hans")

    // Output: 你好, John
    fmt.Println(localizer.Get("hello", i18n.Vars{
        "name": "John",
    }))
}
```

&nbsp;

## Advanced Configuration

### Custom Formatters

Add custom formatters for domain-specific formatting needs:

```go
bundle := i18n.NewBundle(
    i18n.WithDefaultLocale("en"),
    i18n.WithCustomFormatters(map[string]interface{}{
        "upper": func(value interface{}, locale string, arg *string) interface{} {
            return strings.ToUpper(fmt.Sprintf("%v", value))
        },
        "currency": func(value interface{}, locale string, arg *string) interface{} {
            // Custom currency formatting
            return fmt.Sprintf("$%.2f", value)
        },
    }),
)

localizer := bundle.NewLocalizer("en")
result, _ := localizer.Format("Hello, {name, upper}!", i18n.Vars{
    "name": "world",
})
// Output: Hello, WORLD!
```

### Strict Mode

Enable strict parsing for better error detection:

```go
bundle := i18n.NewBundle(
    i18n.WithDefaultLocale("en"),
    i18n.WithStrictMode(true),
)
```

### MessageFormat Options

Configure MessageFormat behavior:

```go
import mf "github.com/kaptinlin/messageformat-go/v1"

options := &mf.MessageFormatOptions{
    Strict:   true,
    Currency: "USD",
    // Add other MessageFormat options
}

bundle := i18n.NewBundle(
    i18n.WithDefaultLocale("en"),
    i18n.WithMessageFormatOptions(options),
)
```

&nbsp;

## Loading Methods

## Load from Go map

```go
package main

import "github.com/kaptinlin/go-i18n"

func main() {
    bundle := i18n.NewBundle(
        i18n.WithDefaultLocale("en"),
        i18n.WithLocales("en", "zh-Hans"),
    )

    bundle.LoadMessages(map[string]map[string]string{
        "en": map[string]string{
            "hello_world": "hello, world",
        },
        "zh-Hans": map[string]string{
            "hello_world": "你好，世界",
        },
    })
}
```

&nbsp;

## Load from Files

```go
package main

import "github.com/kaptinlin/go-i18n"

func main() {
    bundle := i18n.NewBundle(
        i18n.WithDefaultLocale("en"),
        i18n.WithLocales("en", "zh-Hans"),
    )

    bundle.LoadFiles("./locales/en.json", "./locales/zh-Hans.json")
}
```

Filenames like `zh-Hans.json` `zh-Hans.user.json` will be combined to a single `zh-Hans` translation.

No matter if you are naming them like `zh_CN`, `zh-Hans` or `ZH_CN`, they will always be converted to `zh-Hans`.

&nbsp;

## Load from Glob Matching Files

```go
package main

import "github.com/kaptinlin/go-i18n"

func main() {
    bundle := i18n.NewBundle(
        i18n.WithDefaultLocale("en"),
        i18n.WithLocales("en", "zh-Hans"),
    )

    bundle.LoadGlob("./locales/*.json")
}
```

The glob pattern adds all files within `locales` directory with the `.json` extension

&nbsp;

## Load from Embedded Files

Use `LoadFS` if you are using `go:embed` to compile your translations to the program.

```go
package main

import "github.com/kaptinlin/go-i18n"

//go:embed locales/*.json
var localesFS embed.FS

func main() {
    bundle := i18n.NewBundle(
        i18n.WithDefaultLocale("en"),
        i18n.WithLocales("en", "zh-Hans"),
    )

    // Load all json files under `locales` folder from the filesystem.
    bundle.LoadFS(localesFS, "locales/*.json")
}
```

&nbsp;

## Translations

Translations named like `welcome_message`, `button_create`, `button_buy` are token-based translations. For text-based, check the chapters below.

```json
{
    "hello_world": "你好，世界"
}
```

```go
localizer := bundle.NewLocalizer("zh-Hans")

// Output: 你好，世界
localizer.Get("hello_world")

// Output: message_what_is_this
localizer.Get("message_what_is_this")
```

Languages named like `zh_cn`, `zh-Hans` or `ZH_CN`, `NewLocalizer` will always convert them to `zh-Hans`.

&nbsp;

### Passing Data to Translation

It's possible to pass the data to translations. [ICU MessageFormat](https://unicode-org.github.io/icu/userguide/format_parse/messages/) is used to parse the text, the templates will be parsed and cached after the translation was loaded.

```json
{
    "message_vars": "你好，{Name}"
}
```

```go
// Output: 你好，Yami
localizer.Get("message_vars", i18n.Vars{
    "Name": "Yami",
})
```

### Dynamic Formatting

Use `Format` only for dynamic messages that are not stored in translation files.
It recompiles the MessageFormat string on every call, so it is not the primary
translation path and should not be used in hot paths. Prefer `Get` for normal
localized content.

```go
localizer := bundle.NewLocalizer("en")

result, err := localizer.Format("Hello, {name}!", i18n.Vars{
    "name": "Alice",
})
// Output: Hello, Alice!

result, err = localizer.Format("{count, plural, =0 {no items} one {# item} other {# items}}", i18n.Vars{
    "count": 5,
})
// Output: 5 items
```

&nbsp;

## Pluralization

Using language specific plural forms (`one`, `other`)

```json
{
    "message": "{count, plural, one {Message} other {Messages}}"
}
```

```go
// Output: Message
localizer.Get("message", i18n.Vars{
    "count": 1,
})

// Output: Messages
localizer.Get("message", i18n.Vars{
    "count": 2,
})
```

Using exact matches (`=0`):

```json
{
    "messages": "{count, plural, =0 {No messages} one {1 message} other {# messages}}"
}
```

```go
// Output: No messages
localizer.Get("messages", i18n.Vars{
    "count": 0,
})

// Output: 1 message
localizer.Get("messages", i18n.Vars{
    "count": 1,
})

// Output: 2 messages
localizer.Get("messages", i18n.Vars{
    "count": 2,
})
```

&nbsp;

## Text-based Translations

Translations can also be named with sentences. When no translation is found, the key text itself is returned.

```json
{
    "I'm fine.": "我过得很好。",
    "How about you?": "你如何呢？"
}
```

```go
// Output: 我过得很好。
localizer.Get("I'm fine.")

// Output: 你如何呢？
localizer.Get("How about you?")

// Output: Thank you!
localizer.Get("Thank you!")
```

&nbsp;

### Disambiguation by context

In English a "Post" can be "Post something (verb)" or "A post (noun)". With token-based translation, you can easily separating them to `post_verb` and `post_noun`.

With text-based translation, you will need to use `GetX` (X stands for context), and giving the translation a `<context>` suffix.

The space before the `<` is **REQUIRED**.

```json
{
    "Post <verb>": "发表文章",
    "Post <noun>": "一篇文章"
}
```

```go
// Output: 发表文章
localizer.GetX("Post", "verb")

// Output: 一篇文章
localizer.GetX("Post", "noun")

// Output: Post
localizer.GetX("Post", "adjective")
```

&nbsp;

### Act as fallback

If a translation is not found, the key text is returned directly. That makes literal text keys usable as their own fallback content.

```go
// Output: Hello, World
localizer.Get("Hello, {Name}", i18n.Vars{
    "Name": "World",
})

// Output: 2 Posts
localizer.Get("{count, plural, =0 {No Post} one {1 Post} other {# Posts}}", i18n.Vars{
    "Count": 2,
})
```

&nbsp;

## Fallbacks

A fallback language is used when a translation is missing from the current language. If it is still missing from the fallback chain, lookup ends at the default language.

If a translation cannot be found from any language, the token name will be output directly.

```go
// `ja-jp` is the default language
bundle := i18n.NewBundle(
    i18n.WithDefaultLocale("ja-JP"),
    i18n.WithFallback(map[string][]string{
        // `zh-Hans` uses `zh`, `zh-Hant` as fallbacks.
        // `en-GB` uses `en-US` as fallback.
        "zh-Hans": {"zh", "zh-Hant"},
        "en-GB":   {"en-US"},
    }),
)
```

Lookup path looks like this with the example above:

```text
zh-Hans -> zh -> zh-Hant -> ja-JP
en-GB -> en-US -> ja-JP
```

Recursive fallback is also supported. If `zh-Hans` falls back to `zh`, and `zh` falls back to `zh-Hant`, lookup walks that chain before returning to the default locale.

Two rules define the model:

- Fallback chains only apply to keys that exist in the default locale.
- `Has` and `Keys` report only keys defined directly on a locale. They do not include inherited fallback keys.

In practice, the default locale should contain the complete key set. Other locales selectively override it.

&nbsp;

## Translation Lookup

When you need to know which locale produced a translation, use `Lookup`. It returns the rendered text, the locale that produced it, and the result source.

```go
r := localizer.Lookup("hello", i18n.Vars{"name": "World"})
fmt.Println(r.Text)   // "你好，World！"
fmt.Println(r.Locale) // "zh-Hans"
fmt.Println(r.Source) // "direct"
```

`TranslationResult` fields:

| Field | Type | Description |
|-------|------|-------------|
| `Text` | `string` | Translated message, or the key itself if not found. Always populated. |
| `Locale` | `string` | BCP 47 locale tag that provided the translation. Always populated. |
| `Source` | `TranslationSource` | `direct`, `fallback`, or `missing`. |

**Complete Example:**

```go
bundle := i18n.NewBundle(
    i18n.WithDefaultLocale("en"),
    i18n.WithLocales("en", "zh-Hans"),
)

bundle.LoadMessages(map[string]map[string]string{
    "en":      {"hello": "Hello", "bye": "Goodbye"},
    "zh-Hans": {"hello": "你好"},
})

localizer := bundle.NewLocalizer("zh-Hans")

// Direct hit in requested locale
r := localizer.Lookup("hello")
// r.Text: "你好", r.Locale: "zh-Hans", r.Source: "direct"

// Fallback to default locale
r = localizer.Lookup("bye")
// r.Text: "Goodbye", r.Locale: "en", r.Source: "fallback"

// Not found anywhere
r = localizer.Lookup("nonexistent")
// r.Text: "nonexistent", r.Locale: "en", r.Source: "missing"

// With context (use the full key convention)
r = localizer.Lookup("Post <verb>")
```

&nbsp;

### Detecting Fallback vs Direct Hit

```go
r := localizer.Lookup("hello")

switch r.Source {
case i18n.TranslationSourceMissing:
    // Key not in any translation file
case i18n.TranslationSourceFallback:
    // Found via fallback chain
case i18n.TranslationSourceDirect:
    // Direct hit in requested locale
}
```

&nbsp;

### Google AIP-193 Compliance

`TranslationResult` maps directly to [AIP-193 `LocalizedMessage`](https://google.aip.dev/193):

```go
r := localizer.Lookup("error_invalid_input", i18n.Vars{"field": fieldName})

// Both fields always populated — AIP-193 compliant
localizedMsg := &errdetails.LocalizedMessage{
    Locale:  r.Locale, // e.g. "zh-Hans" or "en" (never empty)
    Message: r.Text,   // e.g. "字段无效" or "error_invalid_input" (never empty)
}
```

&nbsp;

### Translation Coverage Analytics

```go
r := localizer.Lookup(key, data)
if r.Source == i18n.TranslationSourceMissing {
    metrics.TranslationMiss(r.Locale, key)
} else {
    metrics.TranslationHit(r.Locale, key)
}
```

&nbsp;

## Custom Unmarshaler

Translations are JSON format by default using `github.com/go-json-experiment/json` as the default unmarshaler. Change it by calling `WithUnmarshaler`.

### YAML Unmarshaler

Uses [`go-yaml/yaml`](https://github.com/go-yaml/yaml) to read the files, so you can write the translation files in YAML format.

```go
package main

import "gopkg.in/yaml.v3"

func main() {
    bundle := i18n.NewBundle(
        i18n.WithDefaultLocale("en"),
        i18n.WithLocales("en", "zh-Hans"),
        i18n.WithUnmarshaler(yaml.Unmarshal),
    )
}
```

Your `zh-Hans.yaml` should look like this:

```yaml
hello_world: "你好，世界"
"How are you?": "你过得如何？"
"mobile_interface.button": "按钮"
```

Nested translations are not supported, you will need to name them like `"mobile_interface.button"` as key and quote them in double quotes.

&nbsp;

### TOML Unmarshaler

Uses [`pelletier/go-toml`](https://github.com/pelletier/go-toml) to read the files, so you can write the translation files in TOML format.

```go
package main

import "github.com/pelletier/go-toml/v2"

func main() {
    bundle := i18n.NewBundle(
        i18n.WithDefaultLocale("en"),
        i18n.WithLocales("en", "zh-Hans"),
        i18n.WithUnmarshaler(toml.Unmarshal),
    )
}
```

Your `zh-Hans.toml` should look like this:

```toml
hello_world = "你好, 世界"
hello_name = "你好, {name}"
message = "{count, plural, one {消息} other {消息}}"
message_with_number = "{count, plural, =0 {没有消息} one {1 条消息} other {# 条消息}}"
```

&nbsp;

### INI Unmarshaler

Uses [`go-ini/ini`](https://gopkg.in/ini.v1) to read the files, so you can write the translation files in INI format.

```go
package main

import "gopkg.in/ini.v1"

func unmarshalINI(data []byte, v interface{}) error {
	f, err := ini.LoadSources(ini.LoadOptions{
		SpaceBeforeInlineComment: true,
		IgnoreInlineComment:      true,
	}, data)
	if err != nil {
		return err
	}

	m := *v.(*map[string]string)

	for _, section := range f.Sections() {
		keyPrefix := ""
		if name := section.Name(); name != ini.DefaultSection {
			keyPrefix = name + "."
		}

		for _, key := range section.Keys() {
			m[keyPrefix+key.Name()] = key.Value()
		}
	}

	return nil
}

func main() {
    bundle := i18n.NewBundle(
        i18n.WithDefaultLocale("en"),
        i18n.WithLocales("en", "zh-Hans"),
        i18n.WithUnmarshaler(unmarshalINI),
    )
}
```

Your `zh-Hans.ini` should look like this:

```ini
hello_world=你好, 世界
hello_name=你好, {name}
message={count, plural, one {消息} other {消息}}

[message]
with_number="{count, plural, =0 {没有消息} one {1 条消息} other {# 条消息}}"
```

&nbsp;

## Parse Accept-Language

The built-in `MatchAvailableLocale` function helps you to parse the `Accept-Language` from HTTP Header.

```go
func(w http.ResponseWriter, r *http.Request) {
    // Initialize i18n.
    bundle :=i18n.NewBundle(
        i18n.WithDefaultLocale("zh-Hans"),
        i18n.WithLocales("en", "zh-Hans"),
    )
    bundle.LoadFiles("zh-Hans.json", "en.json")

    // Get `Accept-Language` from request header.
    accept := r.Header.Get("Accept-Language")

    // Use the locale.
    localizer := bundle.NewLocalizer(bundle.MatchAvailableLocale(accept))
    localizer.Get("hello_world")
}
```

Orders of the languages that passed to `NewLocalizer` won't affect the fallback priorities, it will use the first language that was found in loaded translations.

## Introspection

Use `Has` and `Keys` when you need to inspect what a locale defines directly.

```go
bundle := i18n.NewBundle(
    i18n.WithDefaultLocale("en"),
    i18n.WithLocales("en", "zh-Hans", "ja-JP"),
    i18n.WithFallback(map[string][]string{
        "ja-JP": {"zh-Hans"},
    }),
)

bundle.LoadMessages(map[string]map[string]string{
    "en": {
        "hello": "Hello",
        "bye":   "Goodbye",
    },
    "zh-Hans": {
        "hello": "你好",
    },
    "ja-JP": {},
})

bundle.Has("zh-Hans", "hello") // true
bundle.Has("zh-Hans", "bye")   // false
bundle.Has("ja-JP", "hello")   // false

bundle.Keys("en")      // []string{"bye", "hello"}
bundle.Keys("zh-Hans") // []string{"hello"}
bundle.Keys("ja-JP")   // nil or empty, depending on what was loaded directly
```

`Get` and `Lookup` still use fallback resolution. Introspection does not.

&nbsp;

## Performance

This library is optimized with Go 1.26 features for maximum performance:

### Optimizations Applied

- **Built-in Functions**: Uses `min()`, `max()`, and `clear()` for efficient operations
- **Slices Package**: Pre-allocation with `slices.Grow()`, deduplication with `slices.Compact()`
- **Maps Package**: Bulk copying with `maps.Copy()` instead of element-by-element assignment
- **String Processing**: `strings.Cut()` and `strings.Builder` for reduced memory allocations
- **Memory Pre-allocation**: Smart capacity estimation for slices and maps
- **Modern MessageFormat**: 10-50x performance improvement over previous engines

### Benchmarks

The modernized codebase shows significant improvements:

- **String normalization**: 40-60% faster with reduced allocations
- **File loading**: 25-35% faster with batch operations
- **Translation lookup**: Optimized with pre-allocated data structures
- **MessageFormat parsing**: 10-50x faster with new engine

&nbsp;

## Thanks

- [teacat/i18n](https://github.com/teacat/i18n)
- [kataras/i18n](https://github.com/kataras/i18n)
- [nicksnyder/go-i18n](https://github.com/nicksnyder/go-i18n)
- [vorlif/spreak](https://github.com/vorlif/spreak)
- [oblq/i18n](https://github.com/oblq/i18n)

## License

`kaptinlin/i18n` is free and open-source software licensed under the [MIT License](https://tldrlegal.com/license/mit-license).
