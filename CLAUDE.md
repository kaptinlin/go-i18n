# go-i18n

High-performance Go internationalization library with ICU MessageFormat support.
Use this file for development guidance. For usage examples and installation, see [README.md](README.md).

## Project Overview

This package centers on a small public surface:

- `I18n` bundle for loading translations, locale matching, fallbacks, and runtime caches
- `Localizer` for request- or locale-scoped translation
- optional HTTP integration via `middleware/`

Keep the design Localizer-centric. Do not introduce global mutable translator state.

## Commands

Use `task` targets first:

```bash
# Test
task test                # go test -race ./...
task test-verbose
task test-coverage

# Quality
task fmt
task vet
task lint                # golangci-lint + tidy check
task vuln                # govulncheck ./...
task verify              # deps, fmt, vet, lint, test, vuln

# Benchmarks
task bench
```

Direct Go commands are fine for focused work:

```bash
go test ./...
go test -race ./...
go test ./... -run TestName
go test -bench=. -benchmem ./...
```

## Architecture

Key files:

- `i18n.go` — bundle construction, locale matching, fallback state, MessageFormat compilation, introspection APIs
- `localizer.go` — lookup and formatting API (`Get`, `GetX`, `Lookup`, `Format`)
- `loader.go` — `LoadMessages`, `LoadFiles`, `LoadGlob`, `LoadFS`
- `locale.go` — `Accept-Language` matching via `golang.org/x/text/language`
- `detector.go` — request locale detection from query/cookie/header/Accept-Language
- `context.go` — context helpers for `*Localizer`
- `middleware/` — optional stdlib HTTP middleware

Examples live in `examples/`. Keep them simple and user-facing.

## Design Philosophy

- **KISS** — keep the API small: bundle setup, localizer lookup, loader methods, and optional HTTP helpers.
- **DRY** — reuse locale matching and MessageFormat compilation paths rather than building parallel helpers.
- **YAGNI** — do not add framework integrations, validation layers, or merge/load helpers unless the package actually needs them.
- **Precision over cleverness** — locale behavior must come from `golang.org/x/text/language`, not ad hoc string heuristics.
- **Elegance through reduction** — prefer one clear Localizer path over global shortcuts or mutable translators.
- **Never:** accidental complexity, feature gravity, abstraction theater, configurability cope.

## Coding Rules

### Must Follow

- Use Go 1.26 features where they simplify code.
- Follow Google Go Best Practices and Style Decisions.
- Keep public APIs small and composable.
- Return errors; wrap with context using `fmt.Errorf("...: %w", err)`.
- Use functional options for bundle configuration.
- Use `golang.org/x/text/language` for locale parsing and matching.
- Preserve graceful degradation: invalid MessageFormat templates fall back to raw text where the existing API already does so.
- Keep loaded translations effectively immutable after load; only narrowly scoped runtime caches may be synchronized and mutated.
- Sort deterministic outputs such as key lists.

### Forbidden

- No `panic` in production code.
- No global translator state or mutable singleton locale.
- No framework-specific HTTP integrations in the core package; keep integrations optional.
- No simplistic placeholder engines or pluralization helpers that bypass ICU MessageFormat.
- No premature abstractions; do not add helpers for one call site.
- No feature creep beyond the requested API surface.
- No working around dependency bugs — if a bug or limitation is in a dependency library, do NOT bypass it by reimplementing the dependency's functionality. Instead, create a report file in `reports/` (see Dependency Issue Reporting below).

## Error Handling

- Compilation/loading failures should report actionable context.
- Missing translations should preserve current fallback behavior: use fallback locale when available, otherwise return the key text.
- Runtime formatting failures should preserve the existing graceful fallback to raw text unless the API contract explicitly returns an error.

## Testing

- Use table-driven tests for locale matching, fallback selection, and loader behavior.
- Use `testify/assert` and `testify/require`.
- Test through exported behavior, not private fields, unless validating a narrow cache invariant that has no public signal.
- Use `go test -race ./...` for changes touching caches, middleware, or concurrency.
- Add focused tests for:
  - direct hit vs fallback vs miss
  - context-disambiguated keys
  - MessageFormat formatting and graceful fallback
  - loader behavior across map/files/glob/embed
  - detector source priority and middleware request wiring
- Do not over-test trivial getters or duplicate equivalent locale cases.

## Performance

- Pre-compile MessageFormat templates during load.
- Reuse parsed translations and runtime fallback cache; avoid recompiling in hot paths.
- Prefer standard library and Go 1.26 helpers (`slices`, `maps`, `strings.Cut`) when they reduce allocations or simplify logic.
- Benchmark before adding complexity.

## Dependency Issue Reporting

When you encounter a bug, limitation, or unexpected behavior in a dependency library:

1. **Do NOT** work around it by reimplementing the dependency's functionality
2. **Do NOT** skip or ignore the dependency and write your own version
3. **Do** create a report file: `reports/<dependency-name>.md`
4. **Do** include in the report:
   - Dependency name and version
   - Problem description (what went wrong)
   - Trigger scenario (what you were doing when you hit it)
   - Expected behavior vs actual behavior
   - Relevant error messages or stack traces
   - Workaround suggestion (if any, without implementing it)
5. **Do** continue with other tasks that don't depend on the broken functionality

The `reports/` directory is checked by team members after each work cycle. Reports are routed to the appropriate dependency maintainer for resolution.

## Dependencies

Core dependencies:

- `github.com/kaptinlin/messageformat-go/v1` — ICU MessageFormat engine
- `golang.org/x/text/language` — locale parsing and matching
- `github.com/go-json-experiment/json` — default JSON unmarshaler

Optional format support:

- `github.com/goccy/go-yaml`
- `github.com/pelletier/go-toml/v2`
- `gopkg.in/ini.v1`

Testing:

- `github.com/stretchr/testify`

## Agent Skills

Skills are sourced from `.claude/skills` (symlinked to `.agents/skills`). Use the package-local skills when they match the task.

| Skill | Use when |
|---|---|
| `agent-md-writing` | Regenerating or tightening `CLAUDE.md` / `AGENTS.md`. |
| `code-simplifying` | Simplifying recently changed code without changing behavior. |
| `committing` | Preparing a conventional commit for this package. |
| `dependency-selecting` | Choosing or reviewing external libraries. |
| `go-best-practices` | Checking Go API, naming, error, and interface decisions. |
| `golangci-linting` / `linting` | Configuring linting or fixing lint failures. |
| `modernizing` | Adopting newer Go features safely. |
| `readme-writing` | Updating README content for users. |
| `releasing` | Preparing release/version workflow updates. |
| `testing` | Designing additional tests, benchmarks, or edge-case coverage. |
| `tdd-planning` | Planning implementation in strict TDD slices. |
| `tdd-implementing` | Implementing code through red-green-refactor cycles. |

When a more specialized package skill exists, prefer it over a generic repository-wide skill.
