# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Testing
- `go test -race ./...` - Run tests with race detection
- `make test` - Run tests using the Makefile

### Linting & Code Quality
- `make lint` - Run golangci-lint and tidy checks
- `make golangci-lint` - Run golangci-lint specifically  
- `make tidy-lint` - Check that go.mod/go.sum are tidy
- `make all` - Run both lint and test

### Building
- Standard Go commands: `go build`, `go install`
- Examples can be run with `go run` from their respective directories

## Architecture

This is a Go internationalization (i18n) library that provides localization support with the following key components:

### Core Structure
- **I18n (Bundle)**: Main internationalization core that manages translations, locales, and configuration
  - Handles multiple locales with fallback support
  - Supports custom unmarshalers (JSON, YAML, TOML, INI)
  - Uses golang.org/x/text for language matching and parsing
  
- **Localizer**: Per-locale translator that provides translation methods
  - `Get()` for token-based translations
  - `GetX()` for context-disambiguated translations  
  - `Getf()` for sprintf-style formatting

### Translation System
- **Token-based**: Keys like `hello_world`, `button_create`
- **Text-based**: Full sentences that act as fallbacks when translation missing
- **ICU MessageFormat**: Full support for pluralization, variables, and complex formatting
- **Context support**: Disambiguate translations with `<context>` suffix

### File Loading
- **LoadFiles()**: Load specific translation files
- **LoadGlob()**: Load files matching glob patterns  
- **LoadFS()**: Load from embedded filesystems (go:embed)
- **LoadMessages()**: Load from Go maps

### Language Features
- Language normalization (converts `zh_CN`, `zh-Hans`, etc. to standard forms)
- Fallback chains with recursive support
- Accept-Language header parsing
- Language confidence matching

## Key Files
- `i18n.go` - Core Bundle/I18n struct and initialization
- `localizer.go` - Localizer with translation methods
- `loader.go` - File loading functionality
- `locale.go` - Language/locale utilities
- `types.go` - Type definitions (Vars map)

## Configuration
- Uses golangci-lint with extensive linters enabled
- Go 1.23.0+ required
- Test files exclude some linters (gochecknoglobals, gosec, funlen, etc.)