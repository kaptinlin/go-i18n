package i18n

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/go-json-experiment/json"
	mf "github.com/kaptinlin/messageformat-go/v1"
	"golang.org/x/text/language"
)

// Errors related to translation loading and compilation.
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
	parsedTranslations        map[string]map[string]*parsedTranslation
	runtimeParsedTranslations map[string]*parsedTranslation
	runtimeTranslationsMu     sync.RWMutex
	mfOptions                 *mf.MessageFormatOptions
}

// parsedTranslation holds a pre-compiled translation with its locale, name,
// original text, and an optional compiled MessageFormat function.
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
		i.fallbacks = f
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
		i.mfOptions = opts
	}
}

// WithCustomFormatters adds custom formatters for MessageFormat.
// Creates a new options struct if none exists.
func WithCustomFormatters(formatters map[string]any) Option {
	return func(i *I18n) {
		if i.mfOptions == nil {
			i.mfOptions = &mf.MessageFormatOptions{}
		}
		i.mfOptions.CustomFormatters = formatters
	}
}

// WithStrictMode enables strict parsing mode for MessageFormat.
// Creates a new options struct if none exists.
func WithStrictMode(strict bool) Option {
	return func(i *I18n) {
		if i.mfOptions == nil {
			i.mfOptions = &mf.MessageFormatOptions{}
		}
		i.mfOptions.Strict = strict
	}
}

// NewBundle creates a new internationalization bundle with the given options.
// If no default locale is set, the first locale from [WithLocales] is used;
// if no locales are configured, English is used as the default.
func NewBundle(options ...Option) *I18n {
	i := &I18n{
		unmarshaler:               func(data []byte, v any) error { return json.Unmarshal(data, v) },
		fallbacks:                 make(map[string][]string),
		runtimeParsedTranslations: make(map[string]*parsedTranslation),
		parsedTranslations:        make(map[string]map[string]*parsedTranslation),
	}
	for _, o := range options {
		o(i)
	}
	if i.defaultLanguage == language.Und {
		if len(i.languages) == 0 {
			i.defaultLanguage = language.English
		} else {
			i.defaultLanguage = i.languages[0]
		}
		i.defaultLocale = i.defaultLanguage.String()
	}
	i.ensureDefaultLanguageFirst()
	i.languageMatcher = language.NewMatcher(i.languages)
	return i
}

// SupportedLanguages returns all language tags supported by this bundle.
func (i *I18n) SupportedLanguages() []language.Tag {
	return i.languages
}

// HasTranslation reports whether key exists for locale in the loaded bundle state.
func (i *I18n) HasTranslation(locale, key string) bool {
	resolved, ok := i.resolveLoadedLocale(locale)
	if !ok {
		return false
	}
	translations := i.parsedTranslations[resolved]
	_, ok = translations[key]
	return ok
}

// Keys returns the sorted translation keys for locale in the loaded bundle state.
func (i *I18n) Keys(locale string) []string {
	resolved, ok := i.resolveLoadedLocale(locale)
	if !ok {
		return nil
	}
	translations := i.parsedTranslations[resolved]
	keys := make([]string, 0, len(translations))
	for key := range translations {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func (i *I18n) resolveLoadedLocale(locale string) (string, bool) {
	if locale == "" {
		return "", false
	}
	if matched := i.matchExactLocale(locale); matched != "" {
		_, ok := i.parsedTranslations[matched]
		return matched, ok
	}

	tag, err := language.Parse(locale)
	if err != nil || tag == language.Und || !i.IsLanguageSupported(tag) {
		return "", false
	}

	matched := i.resolveLocale(locale)
	_, ok := i.parsedTranslations[matched]
	return matched, ok
}

// ensureDefaultLanguageFirst ensures the default language is the first element
// in the languages slice, adding it if absent or moving it to the front.
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

// matchExactLocale returns the string form of the supported locale that
// exactly matches the given locale, or an empty string if none matches.
func (i *I18n) matchExactLocale(locale string) string {
	_, idx, conf := i.languageMatcher.Match(language.Make(locale))
	if conf == language.Exact {
		return i.languages[idx].String()
	}
	return ""
}

func (i *I18n) resolveLocale(locale string) string {
	if locale == "" {
		return i.defaultLocale
	}

	_, idx, conf := i.languageMatcher.Match(language.Make(locale))
	if conf == language.No {
		return i.defaultLocale
	}

	matched := i.languages[idx].String()
	if _, ok := i.parsedTranslations[matched]; ok {
		return matched
	}
	return i.defaultLocale
}

// IsLanguageSupported reports whether lang can be matched to a supported locale.
// Languages not in SupportedLanguages may still match through the language matcher.
func (i *I18n) IsLanguageSupported(lang language.Tag) bool {
	_, _, conf := i.languageMatcher.Match(lang)
	return conf > language.No
}

// NewLocalizer creates a Localizer for the first matching locale from
// locales. If none match, the default locale is used.
func (i *I18n) NewLocalizer(locales ...string) *Localizer {
	for _, loc := range locales {
		matched := i.matchExactLocale(loc)
		if matched == "" {
			continue
		}
		if _, ok := i.parsedTranslations[matched]; ok {
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
	if idx := strings.LastIndex(v, " <"); idx != -1 && strings.HasSuffix(v, ">") {
		return v[:idx]
	}
	return v
}

// parseTranslation compiles a translation text into a parsedTranslation.
// Returns [ErrMessageFormatCompilation] if MessageFormat compilation fails,
// along with a parsedTranslation that contains the raw text (usable without formatting).
func (i *I18n) parseTranslation(locale, name, text string) (*parsedTranslation, error) {
	pt := &parsedTranslation{
		name:   name,
		locale: locale,
		text:   text,
	}

	base, _ := language.MustParse(locale).Base()

	formatter, err := mf.New(base.String(), i.mfOptions)
	if err != nil {
		return pt, fmt.Errorf("%w: create formatter: %w", ErrMessageFormatCompilation, err)
	}

	compiled, err := formatter.Compile(text)
	if err != nil {
		return pt, fmt.Errorf("%w: %w", ErrMessageFormatCompilation, err)
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

	if pt = i.runtimeParsedTranslations[name]; pt != nil {
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
			if best := i.lookupBestFallback(locale, defTrans.name); best != nil {
				i.parsedTranslations[locale][defTrans.name] = best
			}
		}
	}
}

// lookupBestFallback finds the best fallback translation for a given locale and
// translation name by traversing the fallback chain.
func (i *I18n) lookupBestFallback(locale, name string) *parsedTranslation {
	return i.lookupFallback(locale, name, make(map[string]struct{}))
}

// lookupFallback recursively searches the fallback chain for a translation.
// The visited set prevents infinite recursion from circular fallback configs.
func (i *I18n) lookupFallback(locale, name string, visited map[string]struct{}) *parsedTranslation {
	if _, ok := visited[locale]; ok {
		return nil
	}
	visited[locale] = struct{}{}

	chain, ok := i.fallbacks[locale]
	if !ok {
		return i.parsedTranslations[i.defaultLocale][name]
	}
	for _, fb := range chain {
		if v, ok := i.parsedTranslations[fb][name]; ok {
			return v
		}
		if found := i.lookupFallback(fb, name, visited); found != nil {
			return found
		}
	}
	return i.parsedTranslations[i.defaultLocale][name]
}
