# go-i18n Specification

## Overview

go-i18n is a Go library for constructing one validated translation bundle,
loading flat locale catalogs, matching runtime locale preferences, and rendering
localized strings with ICU MessageFormat support.

The library is intentionally small. `I18n` owns shared catalog state and locale
rules. `Localizer` owns one matched-locale view over that state. Optional
`net/http` middleware only connects request detection to a request context.

> **Why**: Translation state and request locale choice have different
> lifetimes. Keeping the bundle shared and the localizer cheap makes the common
> path understandable without global mutable state.
>
> **Rejected**: Global translator singletons, request-scoped bundle mutation,
> framework-specific core integrations, and compatibility constructors that
> preserve invalid configuration silently.

## Concept Model

### Bundle

- **Definition**: The validated shared object returned by `NewBundle`.
- **Owns**: supported locales, default locale, fallback rules, directly loaded
  catalog entries, MessageFormat configuration, file unmarshaling, and runtime
  missing-key cache.
- **Lifecycle**: constructed once during application setup, loaded through
  `Load*` methods, then shared by localizers.
- **Invariants**: locale-bearing construction input is valid before a bundle is
  returned; directly loaded catalog state is replaced only after a complete load
  succeeds.

### Catalog

- **Definition**: Directly loaded translations keyed by canonical BCP 47 locale
  and translation key.
- **Includes**: entries from `LoadMessages`, `LoadFiles`, `LoadGlob`, and
  `LoadFS`.
- **Excludes**: fallback-populated copies, nested translation trees, live reload
  policy, and persistence policy.
- **Lifecycle**: each successful load merges staged entries into a cloned
  direct catalog and commits the whole catalog replacement atomically.
- **Invariants**: failed loads leave the previous direct catalog untouched.

### Localizer

- **Definition**: A locale-scoped view over a bundle.
- **Owns**: the matched locale selected from caller preferences.
- **Lifecycle**: created with `NewLocalizer` for requests, jobs, or
  locale-specific work.
- **Invariants**: matching uses supported locales, not whether a locale has
  already loaded translations.

### Lookup Result

- **Definition**: The diagnostic result returned by `Localizer.Lookup`.
- **Fields**:
  - `Text`: rendered translation text, or the key text on miss.
  - `Template`: resolved loaded template, meaning the raw MessageFormat text
    from the direct or fallback catalog entry that supplied `Text`.
  - `MatchedLocale`: the locale selected for the localizer.
  - `CatalogLocale`: the loaded catalog locale that supplied `Text`.
  - `Source`: `direct`, `fallback`, or `missing`.
- **Invariants**: missing results have an empty `CatalogLocale`; direct results
  use the matched locale as catalog locale; fallback results keep matched and
  catalog locale distinct when fallback supplied the text; `Template` is empty
  when `Source` is `missing`.

## Public API Contracts

### Construction

- `NewBundle(options ...Option) (*I18n, error)` is the only bundle constructor.
- `WithDefaultLocale`, `WithLocales`, and `WithFallback` participate in a single
  locale validation contract.
- Invalid or empty locale tags are construction errors.
- Fallback keys and fallback values must resolve exactly to configured
  supported locales.
- If no default locale is supplied, the first supported locale is the default.
  If no locales are supplied, English is the default supported locale.
- `Option` configures construction state; it must not mutate a partially
  constructed bundle.

> **Why**: Locale configuration is startup trust data. Returning a bundle after
> repairing or ignoring bad locale input makes later translation loss look like
> runtime behavior instead of setup failure.
>
> **Rejected**: parallel constructors, best-effort invalid locale filtering,
> and option callbacks that mutate live bundle state during construction.

### Loading

- `LoadMessages`, `LoadFiles`, `LoadGlob`, and `LoadFS` are the only catalog
  loading paths.
- Each load validates every incoming locale before committing any catalog
  changes.
- Each incoming message is compiled before commit.
- File names map to locales through the package filename normalization rule;
  the resulting locale must still be valid and configured.
- Multiple successful loads may add or replace direct entries for a locale.

> **Why**: A failed startup load must not poison the bundle with half of a new
> catalog. Loading either teaches the bundle a complete staged change or teaches
> it nothing.
>
> **Rejected**: silently skipping unknown locale maps or files, in-place repair
> of live catalog maps during parse loops, and a second loader model for merged
> fallback catalogs.

### Locale Matching

- Locale parsing and matching must flow through `golang.org/x/text/language`.
- Construction and loading are strict trust boundaries.
- Runtime locale preferences are forgiving: `NewLocalizer`,
  `MatchAvailableLocale`, and `Detector` return the best supported match or the
  default locale.
- Explicit request sources may be ignored when empty, invalid, or unsupported,
  allowing later detector sources to win.

> **Why**: Setup data should fail loudly, while request data is user input and
> should degrade to the best available locale.
>
> **Rejected**: hand-written locale prefix matching, string-only normalization
> rules, and tying runtime locale selection to whether translations were loaded
> for that locale.

### Fallback Resolution

- The direct catalog is the only stored catalog truth.
- Fallbacks are resolved at lookup time by walking configured fallback chains
  and trying the default locale last.
- Fallback traversal must be cycle-safe.
- `Has` and `Keys` inspect only direct catalog entries.
- `Localizer.GetTemplate` resolves direct and fallback catalog entries like
  `Get`, but returns the raw loaded template without formatting.
- Missing keys return the key text after context suffix trimming where
  applicable.

> **Why**: Fallback is a rule, not loaded data. Keeping fallback out of catalog
> storage makes lookup, diagnostics, and introspection tell the same story.
>
> **Rejected**: fallback-populated catalog maps, inherited keys in `Has` or
> `Keys`, and fallback behavior that changes localizer matching.

### Formatting

- Loaded messages are compiled during load with the configured MessageFormat
  settings.
- `Localizer.Get` and `Localizer.Lookup` format compiled translations and return
  raw message text when runtime formatting fails.
- `Localizer.GetTemplate` does not format; it returns only the resolved loaded
  template and reports false when the result would be a missing-key fallback.
- `Localizer.Format` is for dynamic messages outside the catalog and returns
  compile or runtime errors to the caller.
- Missing-key text may be compiled and cached only as a runtime fallback
  template.

> **Why**: Stored translations are the hot path and should pay compile cost at
> load time. Dynamic messages are caller-owned and can report errors directly.
>
> **Rejected**: custom placeholder engines, custom pluralization logic, and
> recompiling loaded catalog messages on every lookup.

### HTTP Boundary

- `Detector` resolves locale from query, cookie, explicit header, and
  `Accept-Language` sources according to configured priority.
- `middleware.HTTPMiddleware` creates a request-scoped localizer and stores it
  in context.
- The core package must not own sessions, profiles, persistence, or
  framework-specific adapters.

> **Why**: HTTP integration is transport glue. Locale policy beyond request
> detection belongs to the application.
>
> **Rejected**: framework adapters in the core package, session preference
> stores, and global request locale state.

## Failure Semantics

- Construction failure returns `nil, error` and creates no usable bundle.
- Load failure returns an error and leaves the previous direct catalog intact.
- Read, glob, unmarshal, locale, and MessageFormat compilation errors must
  include actionable context such as file path, locale, or key.
- Lookup misses are not errors; they return key text with `Source` set to
  `missing`, empty `CatalogLocale`, and empty `Template`.
- Runtime formatting failure in `Get` or `Lookup` returns raw catalog text.
- Runtime formatting failure in `Format` returns an error.

## Forbidden

- Do not add global mutable translator state. Use `*I18n` plus derived
  `*Localizer` values.
- Do not add alternate constructors or compatibility shims that bypass
  construction errors.
- Do not silently ignore invalid construction locales or unconfigured catalog
  locales.
- Do not store fallback-populated translations as catalog state.
- Do not make `Has` or `Keys` include inherited fallback keys.
- Do not bypass ICU MessageFormat with private interpolation, pluralization, or
  selector logic.
- Do not add framework-specific middleware to the core package.
- Do not add nested catalog shape unless a separate accepted contract replaces
  the flat `map[string]string` loader model.
- Do not work around dependency bugs inline; record dependency issues under
  `reports/` and keep unrelated work moving.

## Acceptance Criteria

- `NewBundle` rejects invalid default locales, supported locales, fallback keys,
  and fallback values.
- `LoadMessages`, `LoadFiles`, `LoadGlob`, and `LoadFS` reject invalid or
  unconfigured catalog locales and preserve the previous catalog on failure.
- `NewLocalizer` can select a supported locale even before translations are
  loaded for that locale.
- Direct, fallback, circular fallback, default fallback, and missing lookup
  paths are covered through public behavior tests.
- `Lookup` reports `Template`, `MatchedLocale`, `CatalogLocale`, and `Source`
  according to the lookup result invariants.
- `Has` and `Keys` expose only direct catalog entries.
- Detector and middleware tests prove request locale detection stays at the HTTP
  edge.
- MessageFormat tests prove stored messages compile on load, runtime lookup
  formatting degrades to raw text, and direct `Format` returns errors.
