package i18n

import (
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
type Option func(*I18n)

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
	parsedTranslations        map[string]map[string]*parsedTranslation
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
	return func(i *I18n) {
		if u == nil {
			return
		}
		i.unmarshaler = u
	}
}

// WithFallback configures locale fallback chains. Each key is a locale, and
// its value is an ordered list of fallback locales to try when a translation
// is missing. The default locale is used as the final fallback.
func WithFallback(f map[string][]string) Option {
	return func(i *I18n) {
		if f == nil {
			i.fallbacks = nil
			return
		}

		fallbacks := make(map[string][]string, len(f))
		for locale, chain := range f {
			fallbacks[locale] = slices.Clone(chain)
		}
		i.fallbacks = fallbacks
	}
}

// WithDefaultLocale sets the default locale. This locale is used when no
// translation is found in the requested locale or its fallback chain.
func WithDefaultLocale(locale string) Option {
	return func(i *I18n) {
		i.defaultLanguage = language.Make(locale)
		i.defaultLocale = i.defaultLanguage.String()
	}
}

// WithLocales sets the supported locales for the bundle.
// Invalid locale strings are silently ignored.
func WithLocales(locales ...string) Option {
	return func(i *I18n) {
		tags := make([]language.Tag, 0, len(locales))
		for _, loc := range locales {
			tag, err := language.Parse(loc)
			if err == nil && tag != language.Und {
				tags = append(tags, tag)
			}
		}
		i.languages = tags
	}
}

// NewBundle creates a new internationalization bundle with the given options.
// If no default locale is set, the first locale from [WithLocales] is used;
// if no locales are configured, English is used as the default.
func NewBundle(options ...Option) *I18n {
	i := &I18n{
		unmarshaler:               func(data []byte, v any) error { return json.Unmarshal(data, v) },
		fallbacks:                 make(map[string][]string),
		directTranslations:        make(map[string]map[string]*parsedTranslation),
		runtimeParsedTranslations: make(map[string]*parsedTranslation),
		parsedTranslations:        make(map[string]map[string]*parsedTranslation),
	}
	for _, o := range options {
		o(i)
	}
	if i.defaultLanguage == language.Und {
		i.defaultLanguage = language.English
		if len(i.languages) > 0 {
			i.defaultLanguage = i.languages[0]
		}
		i.defaultLocale = i.defaultLanguage.String()
	}
	i.ensureDefaultLanguageFirst()
	i.languageMatcher = language.NewMatcher(i.languages)
	return i
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
	return i.resolveLocaleForTable(locale, i.parsedTranslations, true)
}

func (i *I18n) ensureDefaultLanguageFirst() {
	if len(i.languages) == 0 {
		i.languages = []language.Tag{i.defaultLanguage}
		return
	}
	if i.languages[0] == i.defaultLanguage {
		return
	}
	if idx := slices.Index(i.languages, i.defaultLanguage); idx > 0 {
		i.languages = slices.Delete(i.languages, idx, idx+1)
	}
	i.languages = slices.Insert(i.languages, 0, i.defaultLanguage)
}

func (i *I18n) matchExactLocale(locale string) string {
	_, idx, conf := i.languageMatcher.Match(language.Make(locale))
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

// formatFallbacks populates missing translations for each locale by looking up
// the best available fallback from the configured fallback chain.
func (i *I18n) formatFallbacks() {
	i.rebuildParsedTranslations()

	names := make(map[string]struct{})
	for _, translations := range i.directTranslations {
		for name := range translations {
			names[name] = struct{}{}
		}
	}

	for locale, trans := range i.parsedTranslations {
		if locale == i.defaultLocale {
			continue
		}
		for name := range names {
			if _, ok := trans[name]; ok {
				continue
			}
			if best := i.lookupFallback(locale, name); best != nil {
				trans[name] = best
			}
		}
	}
}

func (i *I18n) rebuildParsedTranslations() {
	translations := make(map[string]map[string]*parsedTranslation, len(i.directTranslations))
	for locale, direct := range i.directTranslations {
		translations[locale] = maps.Clone(direct)
	}
	i.parsedTranslations = translations
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

	return i.parsedTranslations[i.defaultLocale][name]
}
