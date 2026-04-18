# go-i18n

High-performance Go internationalization library with ICU MessageFormat support.
For internal contracts and design rules, start with [SPECS/00-overview.md](SPECS/00-overview.md).
For installation and user-facing examples, see [README.md](README.md).

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
task vuln
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
- `detector.go` — request locale detection from query, cookie, header, and `Accept-Language`
- `context.go` — context helpers for `*Localizer`
- `middleware/` — optional stdlib HTTP middleware

Examples live in `examples/`. Keep them simple and user-facing.

## Agent Workflow

- Read [SPECS/00-overview.md](SPECS/00-overview.md) before changing APIs or behavior.
- Keep the design Localizer-centric. Do not introduce global mutable translator state.
- Use package-local skills from `.claude/skills` when they match the task.

## SPECS Index

| Spec | Path | Scope |
|---|---|---|
| Overview | `SPECS/00-overview.md` | Package boundary, locale model, fallback rules, architecture, quality constraints |

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
