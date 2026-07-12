package i18n

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/go-json-experiment/json"
	"golang.org/x/text/language"
)

// Unmarshaler unmarshals translation files. Common implementations include
// json.Unmarshal, yaml.Unmarshal, and toml.Unmarshal.
type Unmarshaler func(data []byte, v any) error

// Option configures an [I18n] bundle. See [WithLocales], [WithFallback], and
// [WithUnmarshaler] for available options.
type Option func(*bundleConfig)

type bundleConfig struct {
	locales       []string
	unmarshaler   Unmarshaler
	fallbacks     map[string][]string
	messageFormat messageFormatter
}

// I18n is the main internationalization bundle that manages translations,
// locales, and fallback chains.
type I18n struct {
	defaultLocale      string
	languages          []language.Tag
	unmarshaler        Unmarshaler
	languageMatcher    language.Matcher
	fallbacks          map[string][]string
	catalogMu          sync.RWMutex
	directTranslations map[string]map[string]*parsedTranslation
	messageFormat      messageFormatter
}

type parsedTranslation struct {
	locale string
	name   string
	text   string
	format messageFunction
}

// WithUnmarshaler sets a custom unmarshaler for translation files.
// The default is JSON. Common alternatives include YAML, TOML, and INI.
func WithUnmarshaler(u Unmarshaler) Option {
	return func(cfg *bundleConfig) {
		if u == nil {
			return
		}
		cfg.unmarshaler = u
	}
}

// WithFallback configures locale fallback chains. Each key is a locale, and
// its value is an ordered list of fallback locales to try when a translation
// is missing. The default locale is always tried last and cannot be an
// explicit fallback source or target. NewBundle rejects duplicate canonical
// declarations and fallback cycles.
func WithFallback(f map[string][]string) Option {
	return func(cfg *bundleConfig) {
		if f == nil {
			cfg.fallbacks = nil
			return
		}

		fallbacks := make(map[string][]string, len(f))
		for locale, chain := range f {
			fallbacks[locale] = slices.Clone(chain)
		}
		cfg.fallbacks = fallbacks
	}
}

// WithLocales sets the additional supported locales. The default locale passed
// to [NewBundle] is already supported and must not be repeated here.
func WithLocales(locales ...string) Option {
	return func(cfg *bundleConfig) {
		cfg.locales = slices.Clone(locales)
	}
}

// NewBundle creates a bundle with an explicit default locale. It returns an
// error for invalid locale configuration or a nil option.
func NewBundle(defaultLocale string, options ...Option) (*I18n, error) {
	cfg := newBundleConfig()
	for index, o := range options {
		if o == nil {
			return nil, fmt.Errorf("bundle option %d is nil", index)
		}
		o(&cfg)
	}
	if err := cfg.messageFormat.validate(); err != nil {
		return nil, err
	}

	languages, defaultLanguage, err := buildLanguages(defaultLocale, cfg)
	if err != nil {
		return nil, err
	}

	i := &I18n{
		defaultLocale:      defaultLanguage.String(),
		languages:          languages,
		unmarshaler:        cfg.unmarshaler,
		directTranslations: make(map[string]map[string]*parsedTranslation),
		messageFormat:      cfg.messageFormat,
	}
	i.languageMatcher = language.NewMatcher(i.languages)

	fallbacks, err := i.normalizeFallbacks(cfg.fallbacks)
	if err != nil {
		return nil, err
	}
	i.fallbacks = fallbacks

	return i, nil
}

func newBundleConfig() bundleConfig {
	return bundleConfig{
		unmarshaler: func(data []byte, v any) error { return json.Unmarshal(data, v) },
		fallbacks:   make(map[string][]string),
	}
}

func buildLanguages(defaultLocale string, cfg bundleConfig) ([]language.Tag, language.Tag, error) {
	defaultLanguage, err := parseConfiguredLocale("default", defaultLocale)
	if err != nil {
		return nil, language.Und, err
	}

	languages := make([]language.Tag, 0, len(cfg.locales)+1)
	languages = append(languages, defaultLanguage)
	sources := map[string]string{defaultLanguage.String(): defaultLocale}
	for _, loc := range cfg.locales {
		tag, err := parseConfiguredLocale("supported", loc)
		if err != nil {
			return nil, language.Und, err
		}
		canonical := tag.String()
		if previous, ok := sources[canonical]; ok {
			return nil, language.Und, fmt.Errorf(
				"locales %q and %q resolve to locale %q",
				previous, loc, canonical,
			)
		}
		sources[canonical] = loc
		languages = append(languages, tag)
	}

	return languages, defaultLanguage, nil
}

// SupportedLocales returns the configured locale tags for this bundle.
func (i *I18n) SupportedLocales() []language.Tag {
	return slices.Clone(i.languages)
}

// Has reports whether key is defined directly for locale in the loaded bundle
// state. Fallback-populated keys are not included.
func (i *I18n) Has(locale, key string) bool {
	translations := i.catalogSnapshot()
	resolved, ok := i.resolveLocaleForTable(locale, translations, false)
	if !ok {
		return false
	}
	_, ok = translations[resolved][key]
	return ok
}

// Keys returns the sorted keys defined directly for locale in the loaded
// bundle state. Fallback-populated keys are not included.
func (i *I18n) Keys(locale string) []string {
	translations := i.catalogSnapshot()
	resolved, ok := i.resolveLocaleForTable(locale, translations, false)
	if !ok {
		return nil
	}
	return slices.Sorted(maps.Keys(translations[resolved]))
}

func (i *I18n) catalogSnapshot() map[string]map[string]*parsedTranslation {
	i.catalogMu.RLock()
	translations := i.directTranslations
	i.catalogMu.RUnlock()
	return translations
}

func (i *I18n) resolveLocaleForTable(
	locale string,
	translations map[string]map[string]*parsedTranslation,
	allowDefault bool,
) (string, bool) {
	if locale == "" {
		return "", false
	}
	if matched := i.matchExactLocale(locale); matched != "" {
		_, ok := translations[matched]
		return matched, ok
	}

	tag, err := language.Parse(locale)
	if err != nil || tag == language.Und {
		return "", false
	}

	_, idx, conf := i.languageMatcher.Match(tag)
	if conf == language.No {
		return "", false
	}

	matched := i.languages[idx].String()
	if _, ok := translations[matched]; ok {
		return matched, true
	}
	if !allowDefault {
		return "", false
	}

	_, ok := translations[i.defaultLocale]
	return i.defaultLocale, ok
}

func (i *I18n) resolveLocalizedLocale(locale string) (string, bool) {
	if locale == "" {
		return "", false
	}

	tag, err := language.Parse(locale)
	if err != nil || tag == language.Und {
		return "", false
	}

	_, idx, conf := i.languageMatcher.Match(tag)
	if conf == language.No {
		return "", false
	}

	return i.languages[idx].String(), true
}

func (i *I18n) matchExactLocale(locale string) string {
	tag, err := language.Parse(locale)
	if err != nil || tag == language.Und {
		return ""
	}

	_, idx, conf := i.languageMatcher.Match(tag)
	if conf == language.Exact {
		return i.languages[idx].String()
	}
	return ""
}

// IsLanguageSupported reports whether lang can be matched to a configured locale.
// Languages not in SupportedLocales may still match through the language matcher.
func (i *I18n) IsLanguageSupported(lang language.Tag) bool {
	_, _, conf := i.languageMatcher.Match(lang)
	return conf > language.No
}

// NewLocalizer creates a Localizer for the first matching locale from
// locales. If none match, the default locale is used.
func (i *I18n) NewLocalizer(locales ...string) *Localizer {
	locale := i.defaultLocale
	for _, loc := range locales {
		if matched, ok := i.resolveLocalizedLocale(loc); ok {
			locale = matched
			break
		}
	}
	return &Localizer{
		bundle: i,
		locale: locale,
	}
}

// trimContext removes the trailing context suffix (e.g., " <verb>") from a
// translation key, returning the base key.
func trimContext(v string) string {
	trimmed, ok := strings.CutSuffix(v, ">")
	if !ok {
		return v
	}
	if idx := strings.LastIndex(trimmed, " <"); idx != -1 {
		return trimmed[:idx]
	}
	return v
}

func (i *I18n) parseTranslation(locale, name, text string) (*parsedTranslation, error) {
	pt := &parsedTranslation{
		name:   name,
		locale: locale,
		text:   text,
	}

	format, err := i.messageFormat.compileTranslation(locale, name, text)
	if err != nil {
		return pt, err
	}

	pt.format = format
	return pt, nil
}

// nameInsensitive normalizes a file name or locale string to a lowercase,
// hyphen-separated form. For example, "zh_CN.music.json" becomes "zh-cn".
func nameInsensitive(v string) string {
	v = filepath.Base(v)
	if before, _, found := strings.Cut(v, "."); found {
		v = before
	}
	return strings.ToLower(strings.ReplaceAll(v, "_", "-"))
}

func (i *I18n) runtimeFallbackTranslation(name string) *parsedTranslation {
	text := trimContext(name)
	pt, err := i.parseTranslation(i.defaultLocale, name, text)
	if err == nil {
		return pt
	}
	return &parsedTranslation{locale: i.defaultLocale, name: name, text: text}
}

func (i *I18n) lookupFallback(
	translations map[string]map[string]*parsedTranslation, locale, name string,
) *parsedTranslation {
	fallbacks := i.fallbacks[locale]
	visited := map[string]struct{}{locale: {}}
	stack := make([]string, 0, len(fallbacks))
	for _, fallback := range slices.Backward(fallbacks) {
		stack = append(stack, fallback)
	}

	for len(stack) > 0 {
		locale := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if _, ok := visited[locale]; ok {
			continue
		}
		visited[locale] = struct{}{}

		if pt, ok := translations[locale][name]; ok {
			return pt
		}
		fallbacks = i.fallbacks[locale]
		for _, fallback := range slices.Backward(fallbacks) {
			stack = append(stack, fallback)
		}
	}

	if _, ok := visited[i.defaultLocale]; ok {
		return nil
	}
	return translations[i.defaultLocale][name]
}

func (i *I18n) resolveLoadLocale(locale string) (string, error) {
	tag, err := parseConfiguredLocale("translation", locale)
	if err != nil {
		return "", err
	}

	_, idx, conf := i.languageMatcher.Match(tag)
	if conf != language.Exact {
		return "", fmt.Errorf("translation locale %q is not configured", locale)
	}
	return i.languages[idx].String(), nil
}

func (i *I18n) normalizeFallbacks(raw map[string][]string) (map[string][]string, error) {
	fallbacks := make(map[string][]string, len(raw))
	sources := make(map[string]string, len(raw))
	for _, locale := range slices.Sorted(maps.Keys(raw)) {
		chain := raw[locale]
		matched, err := i.resolveFallbackLocale("fallback key", locale)
		if err != nil {
			return nil, err
		}
		if matched == i.defaultLocale {
			return nil, fmt.Errorf("default locale %q cannot have explicit fallbacks", i.defaultLocale)
		}
		if previous, ok := sources[matched]; ok {
			return nil, fmt.Errorf(
				"fallback keys %q and %q resolve to locale %q",
				previous, locale, matched,
			)
		}
		sources[matched] = locale

		normalized := make([]string, 0, len(chain))
		seen := make(map[string]struct{}, len(chain))
		for _, fallback := range chain {
			matchedFallback, err := i.resolveFallbackLocale("fallback value", fallback)
			if err != nil {
				return nil, err
			}
			if matchedFallback == i.defaultLocale {
				return nil, fmt.Errorf("default locale %q cannot be an explicit fallback", i.defaultLocale)
			}
			if _, ok := seen[matchedFallback]; ok {
				return nil, fmt.Errorf(
					"fallback locale %q has duplicate target %q",
					matched, matchedFallback,
				)
			}
			seen[matchedFallback] = struct{}{}
			normalized = append(normalized, matchedFallback)
		}
		fallbacks[matched] = normalized
	}
	if err := validateFallbackCycles(fallbacks); err != nil {
		return nil, err
	}
	return fallbacks, nil
}

func validateFallbackCycles(fallbacks map[string][]string) error {
	const (
		visiting = 1
		visited  = 2
	)

	states := make(map[string]int, len(fallbacks))
	path := make([]string, 0, len(fallbacks))
	var visit func(string) error
	visit = func(locale string) error {
		states[locale] = visiting
		path = append(path, locale)
		defer func() { path = path[:len(path)-1] }()

		for _, fallback := range fallbacks[locale] {
			switch states[fallback] {
			case visiting:
				start := slices.Index(path, fallback)
				cycle := append(slices.Clone(path[start:]), fallback)
				return fmt.Errorf("fallback cycle: %s", strings.Join(cycle, " -> "))
			case visited:
				continue
			}
			if err := visit(fallback); err != nil {
				return err
			}
		}
		states[locale] = visited
		return nil
	}

	for _, locale := range slices.Sorted(maps.Keys(fallbacks)) {
		if states[locale] == visited {
			continue
		}
		if err := visit(locale); err != nil {
			return err
		}
	}
	return nil
}

func (i *I18n) resolveFallbackLocale(kind, locale string) (string, error) {
	tag, err := parseConfiguredLocale(kind, locale)
	if err != nil {
		return "", err
	}

	_, idx, conf := i.languageMatcher.Match(tag)
	if conf != language.Exact {
		return "", fmt.Errorf("%s locale %q is not configured", kind, locale)
	}
	return i.languages[idx].String(), nil
}

func parseConfiguredLocale(kind, locale string) (language.Tag, error) {
	if strings.TrimSpace(locale) == "" {
		return language.Und, fmt.Errorf("%s locale is empty", kind)
	}

	tag, err := language.Parse(locale)
	if err != nil || tag == language.Und {
		if err == nil {
			return language.Und, fmt.Errorf("%s locale %q is invalid", kind, locale)
		}
		return language.Und, fmt.Errorf("%s locale %q is invalid: %w", kind, locale, err)
	}
	return tag, nil
}
