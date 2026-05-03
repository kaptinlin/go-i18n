package i18n

import (
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/go-json-experiment/json"
	mf "github.com/kaptinlin/messageformat-go/v1"
	"golang.org/x/text/language"
)

var (
	// ErrMessageFormatCompilation indicates that MessageFormat template compilation failed.
	// The translation text is returned as-is without formatting capabilities.
	ErrMessageFormatCompilation = errors.New("messageformat compilation failed")
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
	mfOptions                 *mf.MessageFormatOptions
}

type parsedTranslation struct {
	locale string
	name   string
	text   string
	format mf.MessageFunction
}

// WithUnmarshaler sets a custom unmarshaler for translation files.
// The default is JSON. Common alternatives include YAML, TOML, and INI.
func WithUnmarshaler(u Unmarshaler) Option {
	return func(i *I18n) {
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

// WithMessageFormatOptions sets MessageFormat options for the bundle.
func WithMessageFormatOptions(opts *mf.MessageFormatOptions) Option {
	return func(i *I18n) {
		if opts == nil {
			i.mfOptions = nil
			return
		}

		i.mfOptions = new(*opts)
		i.mfOptions.CustomFormatters = maps.Clone(opts.CustomFormatters)
	}
}

func (i *I18n) ensureMessageFormatOptions() *mf.MessageFormatOptions {
	if i.mfOptions == nil {
		i.mfOptions = &mf.MessageFormatOptions{}
	}
	return i.mfOptions
}

// WithCustomFormatters adds custom formatters for MessageFormat.
func WithCustomFormatters(formatters map[string]any) Option {
	return func(i *I18n) {
		i.ensureMessageFormatOptions().CustomFormatters = maps.Clone(formatters)
	}
}

// WithStrictMode enables strict parsing mode for MessageFormat.
func WithStrictMode(strict bool) Option {
	return func(i *I18n) {
		i.ensureMessageFormatOptions().Strict = strict
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

func messageFormatBase(locale string) (string, error) {
	tag, err := language.Parse(locale)
	if err != nil {
		return "", fmt.Errorf("parse locale %q: %w", locale, err)
	}
	base, _ := tag.Base()
	return base.String(), nil
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
	for _, loc := range locales {
		if matched, ok := i.resolveLocalizedLocale(loc); ok {
			return &Localizer{
				bundle: i,
				locale: matched,
			}
		}
	}
	return &Localizer{
		bundle: i,
		locale: i.defaultLocale,
	}
}

// trimContext removes the trailing context suffix (e.g., " <verb>") from a
// translation key, returning the base key.
func trimContext(v string) string {
	trimmed, ok := strings.CutSuffix(v, ">")
	if idx := strings.LastIndex(trimmed, " <"); ok && idx != -1 {
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

	base, err := messageFormatBase(locale)
	if err != nil {
		return pt, fmt.Errorf("%w for locale %q key %q: %w", ErrMessageFormatCompilation, locale, name, err)
	}

	formatter, err := mf.New(base, i.mfOptions)
	if err != nil {
		return pt, fmt.Errorf("%w for locale %q key %q: create formatter: %w", ErrMessageFormatCompilation, locale, name, err)
	}

	compiled, err := formatter.Compile(text)
	if err != nil {
		return pt, fmt.Errorf("%w for locale %q key %q: %w", ErrMessageFormatCompilation, locale, name, err)
	}

	pt.format = compiled
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

	pt, _ = i.parseTranslation(i.defaultLocale, name, trimContext(name))
	i.runtimeParsedTranslations[name] = pt
	return pt
}

// formatFallbacks populates missing translations for each locale by looking up
// the best available fallback from the configured fallback chain.
func (i *I18n) formatFallbacks() {
	for _, defTrans := range i.parsedTranslations[i.defaultLocale] {
		for locale, trans := range i.parsedTranslations {
			if locale == i.defaultLocale {
				continue
			}
			if _, ok := trans[defTrans.name]; ok {
				continue
			}
			if best := i.lookupFallback(locale, defTrans.name, make(map[string]struct{})); best != nil {
				i.parsedTranslations[locale][defTrans.name] = best
			}
		}
	}
}

// lookupFallback recursively searches the fallback chain for a translation.
// The visited set prevents infinite recursion from circular fallback configs.
func (i *I18n) lookupFallback(locale, name string, visited map[string]struct{}) *parsedTranslation {
	if _, ok := visited[locale]; ok {
		return nil
	}
	visited[locale] = struct{}{}

	for _, fb := range i.fallbacks[locale] {
		if v, ok := i.parsedTranslations[fb][name]; ok {
			return v
		}
		if found := i.lookupFallback(fb, name, visited); found != nil {
			return found
		}
	}
	return i.parsedTranslations[i.defaultLocale][name]
}
