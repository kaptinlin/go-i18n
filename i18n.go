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

// Option configures an [I18n] bundle. See [WithDefaultLocale],
// [WithLocales], [WithFallback], and [WithUnmarshaler] for available options.
type Option func(*bundleConfig)

type bundleConfig struct {
	defaultLocale    string
	hasDefaultLocale bool
	locales          []string
	unmarshaler      Unmarshaler
	fallbacks        map[string][]string
	messageFormat    messageFormatter
}

// I18n is the main internationalization bundle that manages translations,
// locales, and fallback chains.
type I18n struct {
	defaultLocale             string
	defaultLanguage           language.Tag
	languages                 []language.Tag
	unmarshaler               Unmarshaler
	languageMatcher           language.Matcher
	fallbacks                 map[string][]string
	directTranslations        map[string]map[string]*parsedTranslation
	runtimeParsedTranslations map[string]*parsedTranslation
	runtimeTranslationsMu     sync.RWMutex
	messageFormat             messageFormatter
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
// is missing. The default locale is used as the final fallback.
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

// WithDefaultLocale sets the default locale. This locale is used when no
// translation is found in the requested locale or its fallback chain.
func WithDefaultLocale(locale string) Option {
	return func(cfg *bundleConfig) {
		cfg.defaultLocale = locale
		cfg.hasDefaultLocale = true
	}
}

// WithLocales sets the supported locales for the bundle.
func WithLocales(locales ...string) Option {
	return func(cfg *bundleConfig) {
		cfg.locales = slices.Clone(locales)
	}
}

// NewBundle creates a new internationalization bundle with the given options.
// If no default locale is set, the first locale from [WithLocales] is used;
// if no locales are configured, English is used as the default.
func NewBundle(options ...Option) (*I18n, error) {
	cfg := newBundleConfig()
	for _, o := range options {
		o(&cfg)
	}

	languages, defaultLanguage, err := buildLanguages(cfg)
	if err != nil {
		return nil, err
	}

	i := &I18n{
		defaultLocale:             defaultLanguage.String(),
		defaultLanguage:           defaultLanguage,
		languages:                 languages,
		unmarshaler:               cfg.unmarshaler,
		directTranslations:        make(map[string]map[string]*parsedTranslation),
		runtimeParsedTranslations: make(map[string]*parsedTranslation),
		messageFormat:             cfg.messageFormat,
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

func buildLanguages(cfg bundleConfig) ([]language.Tag, language.Tag, error) {
	languages := make([]language.Tag, 0, len(cfg.locales)+1)
	for _, loc := range cfg.locales {
		tag, err := parseConfiguredLocale("supported", loc)
		if err != nil {
			return nil, language.Und, err
		}
		languages = appendUniqueLanguage(languages, tag)
	}

	defaultLanguage := language.English
	if len(languages) > 0 {
		defaultLanguage = languages[0]
	}
	if cfg.hasDefaultLocale {
		tag, err := parseConfiguredLocale("default", cfg.defaultLocale)
		if err != nil {
			return nil, language.Und, err
		}
		defaultLanguage = tag
	}

	languages = moveLanguageFirst(languages, defaultLanguage)
	if len(languages) == 0 {
		languages = []language.Tag{defaultLanguage}
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
	resolved, ok := i.resolveLocaleForTable(locale, i.directTranslations, false)
	if !ok {
		return false
	}
	translations := i.directTranslations[resolved]
	_, ok = translations[key]
	return ok
}

// Keys returns the sorted keys defined directly for locale in the loaded
// bundle state. Fallback-populated keys are not included.
func (i *I18n) Keys(locale string) []string {
	resolved, ok := i.resolveLocaleForTable(locale, i.directTranslations, false)
	if !ok {
		return nil
	}
	return slices.Sorted(maps.Keys(i.directTranslations[resolved]))
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

func (i *I18n) getRuntimeParsedTranslation(name string) *parsedTranslation {
	i.runtimeTranslationsMu.RLock()
	pt := i.runtimeParsedTranslations[name]
	i.runtimeTranslationsMu.RUnlock()
	if pt != nil {
		return pt
	}

	i.runtimeTranslationsMu.Lock()
	defer i.runtimeTranslationsMu.Unlock()

	pt = i.runtimeParsedTranslations[name]
	if pt != nil {
		return pt
	}

	pt = i.runtimeFallbackTranslation(name)
	i.runtimeParsedTranslations[name] = pt
	return pt
}

func (i *I18n) runtimeFallbackTranslation(name string) *parsedTranslation {
	text := trimContext(name)
	pt, err := i.parseTranslation(i.defaultLocale, name, text)
	if err == nil {
		return pt
	}
	return &parsedTranslation{locale: i.defaultLocale, name: name, text: text}
}

func (i *I18n) lookupFallback(locale, name string) *parsedTranslation {
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

		if pt, ok := i.directTranslations[locale][name]; ok {
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
	return i.directTranslations[i.defaultLocale][name]
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
	for locale, chain := range raw {
		matched, err := i.resolveFallbackLocale("fallback key", locale)
		if err != nil {
			return nil, err
		}

		normalized := make([]string, 0, len(chain))
		for _, fallback := range chain {
			matchedFallback, err := i.resolveFallbackLocale("fallback value", fallback)
			if err != nil {
				return nil, err
			}
			normalized = append(normalized, matchedFallback)
		}
		fallbacks[matched] = normalized
	}
	return fallbacks, nil
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

func appendUniqueLanguage(languages []language.Tag, tag language.Tag) []language.Tag {
	if slices.Contains(languages, tag) {
		return languages
	}
	return append(languages, tag)
}

func moveLanguageFirst(languages []language.Tag, first language.Tag) []language.Tag {
	if len(languages) == 0 {
		return []language.Tag{first}
	}
	if languages[0] == first {
		return languages
	}
	if idx := slices.Index(languages, first); idx > 0 {
		languages = slices.Delete(languages, idx, idx+1)
	}
	return slices.Insert(languages, 0, first)
}
