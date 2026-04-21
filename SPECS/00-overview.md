# go-i18n — Overview

## Overview

go-i18n is a Go library for loading translations, matching locales, and rendering localized strings with ICU MessageFormat support. Its scope is a small public surface: bundle construction, locale-scoped lookup, loader entry points, and optional `net/http` middleware.

The library is not a framework, not a global translator runtime, and not a custom template engine. Applications compose it by constructing one `*I18n` bundle and deriving `*Localizer` values for requests, jobs, or locale-specific work.

## Public Surface

### Bundle and Localizer

- `I18n` owns loaded translations, locale matching, fallback state, and runtime caches.
- `Localizer` performs locale-scoped lookup and formatting through `Get`, `GetX`, `Lookup`, and `Format`.
- `middleware/` is optional integration for stdlib HTTP only.

> **Why**: Bundle construction and locale-scoped lookup have different lifetimes. Sharing immutable translation data through `I18n` and deriving `Localizer` instances keeps the hot path small without introducing global mutable state.
>
> **Rejected**: Global translator singletons, request-agnostic mutable locale state, and framework-specific integrations in the core package.

### Translation and Fallback Model

- Translations load through `LoadMessages`, `LoadFiles`, `LoadGlob`, or `LoadFS`.
- Loaded translations are treated as immutable after load; only narrowly scoped runtime caches may mutate under synchronization.
- Fallback resolution is rooted in the configured default locale.
- `Lookup` reports whether a result was a direct hit, a fallback hit, or a miss.
- `Has` and `Keys` report only keys defined directly on a locale, not inherited fallback keys.

> **Why**: Lookup and introspection serve different needs. Keeping fallback resolution explicit in `Lookup` while limiting `Has` and `Keys` to direct keys preserves deterministic analytics and avoids surprising inherited state in inspection APIs.
>
> **Rejected**: Introspection APIs that silently include inherited keys, and ad hoc merge helpers that create a second translation-loading model.

### Locale Handling

- Locale parsing and matching must use `golang.org/x/text/language`.
- Request detection may combine query, cookie, header, and `Accept-Language` sources, but locale normalization still flows through `language.Tag` parsing and matching.
- Locale behavior must come from BCP 47 semantics, not string heuristics.

> **Why**: Locale matching is correctness-sensitive. Delegating parsing and matching to `golang.org/x/text/language` keeps behavior aligned with standard BCP 47 handling instead of accumulating partial custom logic.
>
> **Rejected**: Hand-rolled locale normalization and string-prefix matching rules.

### Message Formatting

- Localized messages use `github.com/kaptinlin/messageformat-go/v1` as the formatting engine.
- ICU MessageFormat is the only supported path for pluralization and selector behavior.
- Invalid templates and runtime formatting failures fall back to raw text where the existing API already exposes graceful degradation.
- `Format` exists for dynamic messages that are not loaded from translation files; it is not the primary translation path.

> **Why**: ICU MessageFormat provides the required pluralization and selector model, while graceful degradation preserves the library's current behavior for malformed or missing formatting input.
>
> **Rejected**: Simplistic placeholder engines, custom pluralization helpers that bypass ICU MessageFormat, and hard-failing every malformed localized string in APIs that currently degrade gracefully.

## Package Layout

- `i18n.go` — bundle construction, locale matching, fallback state, MessageFormat compilation, introspection APIs
- `localizer.go` — lookup and formatting API
- `loader.go` — `LoadMessages`, `LoadFiles`, `LoadGlob`, `LoadFS`
- `locale.go` — `Accept-Language` matching via `golang.org/x/text/language`
- `detector.go` — request locale detection from query, cookie, header, and `Accept-Language`
- `context.go` — context helpers for `*Localizer`
- `middleware/` — optional stdlib HTTP middleware
- `examples/` — user-facing usage examples

> **Why**: The package layout mirrors the public model: bundle setup, locale detection, loading, and localized lookup are separate concerns with small entry points.
>
> **Rejected**: A monolithic package surface with mixed loading, detection, and transport-specific logic in the same file or type.

## Quality Standards

### Error Handling

- Return errors explicitly and wrap underlying failures with `fmt.Errorf("...: %w", err)`.
- Loading and compilation failures must include actionable context.
- Missing translations fall back through the configured chain and finally return the key text when no translation exists.
- Runtime formatting failures preserve the existing graceful fallback to raw text unless an API explicitly returns an error.

### Testing

- Use table-driven tests and `testify/assert` plus `testify/require`.
- Test exported behavior rather than private fields, except for narrow cache invariants with no public signal.
- Run `go test -race ./...` for changes touching caches, middleware, or concurrency.
- Cover direct hit vs fallback vs miss, context-disambiguated keys, MessageFormat formatting, loader behavior, detector source priority, and middleware wiring.

### Performance

- Pre-compile MessageFormat templates during load.
- Reuse parsed translations and runtime fallback caches.
- Prefer standard library and Go 1.26 helpers when they simplify code or reduce allocations.
- Benchmark before adding complexity to hot paths.

> **Why**: The steady-state cost is translation lookup, not configuration. Work should bias toward predictable lookup behavior and precomputation during load.
>
> **Rejected**: Recompiling templates on hot paths and adding abstractions before benchmarks justify them.

## Dependencies

### Core

- `github.com/kaptinlin/messageformat-go/v1` — ICU MessageFormat engine
- `golang.org/x/text/language` — locale parsing and matching
- `github.com/go-json-experiment/json` — default JSON unmarshaler

### Optional Format Support

- `github.com/goccy/go-yaml`
- `github.com/pelletier/go-toml/v2`
- `gopkg.in/ini.v1`

### Test-Only

- `github.com/stretchr/testify`

## Forbidden

- Do not introduce global translator state. Use `*Localizer` values derived from `*I18n`.
- Do not add framework-specific HTTP integrations to the core package. Keep integrations optional under dedicated packages.
- Do not bypass ICU MessageFormat with custom placeholder or pluralization logic.
- Do not `panic` in production code. Return wrapped errors instead.
- Do not add helpers or configuration layers for one call site. Reuse the existing loader, matcher, and localizer paths.
- Do not work around dependency bugs by reimplementing dependency behavior. Record the problem in `reports/<dependency-name>.md` instead.

## Acceptance Criteria

- [ ] `I18n` and `Localizer` remain the primary public workflow for bundle setup and localized lookup.
- [ ] Locale parsing and matching behavior flows through `golang.org/x/text/language`.
- [ ] Fallback behavior remains explicit and deterministic across direct hits, fallback hits, and misses.
- [ ] Localized formatting continues to use ICU MessageFormat with graceful fallback where the API contract already guarantees it.
- [ ] Loader APIs continue to cover maps, files, glob patterns, and embedded filesystems.
- [ ] Tests continue to cover locale matching, fallback behavior, loading, formatting, detector priority, and middleware wiring.

**Origin:** Migrated from `CLAUDE.md`.
