# go-i18n

Go internationalization library with ICU MessageFormat support, deterministic fallback chains, and optional `net/http` locale detection.
For user-facing installation and examples, see [README.md](README.md).
For internal contracts and design rules, start with [SPECS/00-overview.md](SPECS/00-overview.md).

## Commands

Use `task` targets first:

```bash
# Tests
task test
task test-verbose
task test-coverage
task bench

# Quality
task fmt
task vet
task lint

task vuln
task verify
```

Direct Go commands are fine for focused work:

```bash
go test -race -p 1 ./...
go test ./... -run TestName
go test -bench=. -benchmem ./...
```

## Architecture

Key files and directories:

- `i18n.go` — bundle construction, locale matching, fallback state, MessageFormat compilation, introspection APIs
- `localizer.go` — locale-scoped lookup and formatting APIs
- `loader.go` — `LoadMessages`, `LoadFiles`, `LoadGlob`, `LoadFS`
- `locale.go` — `Accept-Language` matching via `golang.org/x/text/language`
- `detector.go` — request locale detection from query, cookie, header, and `Accept-Language`
- `context.go` — context helpers for `*Localizer`
- `middleware/http.go` — optional stdlib HTTP middleware
- `result.go` — `TranslationResult` and `TranslationSource`
- `examples/` — runnable usage examples

## Design Philosophy

- **KISS** — Keep shared translation state in `I18n` and request-specific work in `Localizer`.
- **SRP** — Keep loading, locale detection, context plumbing, and HTTP middleware in separate entry points.
- **APIs as language** — Prefer small verbs like `LoadFiles`, `LoadFS`, `NewLocalizer`, and `Lookup` over configuration-heavy call sites.
- **Errors as teachers** — Include file names, locales, and keys in loading and formatting failures.
- **Never:** accidental complexity, feature gravity, abstraction theater, configurability cope.

## API Design Principles

- **Progressive Disclosure**: Keep the common path short with `NewBundle`, `Load*`, and `Get`, while leaving lower-level detection and formatting APIs available when needed.
- **Default Passthrough**: Let option zero values preserve the package defaults instead of introducing shadow defaults at call sites.

## Agent Workflow

### Design Phase — Read SPECS First

Before changing APIs, locale behavior, fallback rules, or middleware behavior, read [SPECS/00-overview.md](SPECS/00-overview.md) completely.
Keep the bundle/localizer split, locale model, and fallback semantics aligned with that spec.
If the requested change would alter those contracts, update the spec first or ask the user.

### Implementation Phase — Follow Current Package Patterns

1. Read the adjacent package files and matching tests before editing code.
2. Reuse the existing option-based API shape instead of adding parallel constructors or config layers.
3. Keep request-scoped behavior in `Detector`, context helpers, or `middleware/`, not in global package state.
4. Use package-local skills from `.agents/skills/` when they match the task.

## Dependency Issue Reporting

When you encounter a bug, limitation, or unexpected behavior in a dependency library:

1. Do not work around it by reimplementing the dependency's functionality.
2. Do not skip the dependency and write a private replacement inline.
3. Create a report file at `reports/<dependency-name>.md`.
4. Include the dependency name and version, the trigger scenario, expected behavior, actual behavior, relevant errors, and any non-code workaround suggestion.
5. Continue with unrelated work that does not depend on the broken dependency path.

## SPECS Index

| Spec | Path | Scope |
|---|---|---|
| Overview | `SPECS/00-overview.md` | Package boundary, locale model, fallback rules, architecture, and quality constraints |

## Coding Rules

### Must Follow

- Use Go 1.26.2 features when they simplify code without obscuring behavior.
- Keep locale parsing and matching on `golang.org/x/text/language`.
- Keep `github.com/kaptinlin/messageformat-go/v1` as the MessageFormat dependency; this package currently targets ICU MessageFormat v1 syntax and semantics, not the MessageFormat 2.0 mainline.
- Keep the module graph on a release that still ships the `v1` compatibility package.
- Keep loaded translations effectively immutable after load; only runtime caches may mutate under synchronization.
- Keep translation lookup Localizer-centric; `I18n` owns shared state and locale matching.
- Reuse the existing `Option` helpers for configuration before adding new public knobs.
- Use table-driven tests with `assert` and `require`, and call `t.Parallel()` when the test is safe to parallelize.

### Domain Patterns

See [SPECS/00-overview.md](SPECS/00-overview.md) for the locale model, fallback behavior, quality expectations, and forbidden architectural directions.

### Forbidden

- Do not introduce global mutable translator state.
- Do not add framework-specific HTTP integrations to the core package.
- Do not bypass ICU MessageFormat with custom placeholder or pluralization logic.
- Do not switch to the unversioned `github.com/kaptinlin/messageformat-go` MessageFormat 2.0 API unless the repository has been explicitly migrated off ICU MessageFormat v1 syntax end-to-end.
- Do not encode spec prose as constants, enums, or helpers that no runtime code consumes.
- Do not work around dependency bugs inline; use `reports/<dependency-name>.md` instead.
- Do not `panic` in package code; return wrapped errors instead.

## Testing

- Run `task test` for normal validation and `go test -race -p 1 ./...` for focused race-safe checks.
- Run `task lint` before shipping changes; it includes `golangci-lint` and tidy checks.
- Run `task bench` before claiming a performance improvement.
- Keep docs, examples, and tests aligned when public behavior changes.

## Dependencies

Core dependencies:

- `github.com/kaptinlin/messageformat-go/v1` — ICU MessageFormat engine
- `golang.org/x/text/language` — locale parsing and matching
- `github.com/go-json-experiment/json` — default JSON unmarshaler

Optional format support:

- `github.com/goccy/go-yaml`
- `github.com/pelletier/go-toml/v2`
- `gopkg.in/ini.v1`

## Agent Skills

Package-local skills live under `.agents/skills/`.
Use the most specific matching skill before falling back to a generic repository skill.

| Skill | Use when |
|---|---|
| `go-i18n-localizing` | Adding or reviewing loaders, locale handling, lookup behavior, or middleware-facing localization flows. |
| `library-docs-maintaining` | Refreshing `README.md`, `CLAUDE.md`, and the `AGENTS.md` symlink. |
| `library-specs-maintaining` | Updating `SPECS/` to match current package contracts. |
| `library-test-covering` | Expanding coverage for loaders, lookup, locale detection, or middleware behavior. |
| `library-error-optimizing` | Tightening error messages and error wrapping. |
| `library-panic-optimizing` | Removing panics from package code. |
| `go-best-practices` | Reviewing API shape, naming, interfaces, and Go conventions. |
| `golangci-linting` | Configuring lint rules or fixing lint failures. |
| `agent-md-writing` | Regenerating `CLAUDE.md` and verifying the `AGENTS.md` symlink. |
| `readme-writing` | Updating user-facing README content. |
| `committing` | Preparing a conventional commit for this package. |
