package i18n

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-json-experiment/json"
	mf "github.com/kaptinlin/messageformat-go/v1"
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
	parsedTranslations        map[string]map[string]*parsedTranslation
	runtimeParsedTranslations map[string]*parsedTranslation
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

// WithUnmarshaler replaces the default JSON unmarshaler for translation files.
func WithUnmarshaler(u Unmarshaler) Option {
	return func(b *I18n) {
		b.unmarshaler = u
	}
}

// WithFallback configures locale fallback chains. Each key is a locale, and
// its value is an ordered list of locales to try before the default locale.
func WithFallback(f map[string][]string) Option {
	return func(b *I18n) {
		b.fallbacks = f
	}
}

// WithDefaultLocale sets the default locale, used as the ultimate fallback
// when no translation is found in the requested or fallback locales.
func WithDefaultLocale(locale string) Option {
	return func(b *I18n) {
		b.defaultLanguage = language.Make(locale)
		b.defaultLocale = b.defaultLanguage.String()
	}
}

// WithLocales configures the supported locales for the bundle.
// Invalid locale strings are silently ignored.
func WithLocales(locales ...string) Option {
	return func(b *I18n) {
		tags := make([]language.Tag, 0, len(locales))
		for _, loc := range locales {
			tag, err := language.Parse(loc)
			if err == nil && tag != language.Und {
				tags = append(tags, tag)
			}
		}
		b.languages = tags
	}
}

// WithMessageFormatOptions sets MessageFormat options for the bundle.
func WithMessageFormatOptions(opts *mf.MessageFormatOptions) Option {
	return func(b *I18n) {
		b.mfOptions = opts
	}
}

// WithCustomFormatters sets custom formatters for MessageFormat.
// If no MessageFormat options have been set, a new options struct is created.
func WithCustomFormatters(formatters map[string]any) Option {
	return func(b *I18n) {
		if b.mfOptions == nil {
			b.mfOptions = &mf.MessageFormatOptions{}
		}
		b.mfOptions.CustomFormatters = formatters
	}
}

// WithStrictMode enables or disables strict parsing mode for MessageFormat.
// If no MessageFormat options have been set, a new options struct is created.
func WithStrictMode(strict bool) Option {
	return func(b *I18n) {
		if b.mfOptions == nil {
			b.mfOptions = &mf.MessageFormatOptions{}
		}
		b.mfOptions.Strict = strict
	}
}

// NewBundle creates a new internationalization bundle with the given options.
// If no default locale is set, the first locale from [WithLocales] is used;
// if no locales are configured, English is used as the default.
func NewBundle(options ...Option) *I18n {
	b := &I18n{
		unmarshaler:               func(data []byte, v any) error { return json.Unmarshal(data, v) },
		fallbacks:                 make(map[string][]string),
		runtimeParsedTranslations: make(map[string]*parsedTranslation),
		parsedTranslations:        make(map[string]map[string]*parsedTranslation),
	}
	for _, o := range options {
		o(b)
	}
	if b.defaultLanguage == language.Und {
		switch {
		case len(b.languages) == 0:
			b.defaultLanguage = language.English
		default:
			b.defaultLanguage = b.languages[0]
		}
		b.defaultLocale = b.defaultLanguage.String()
	}
	b.ensureDefaultLanguageFirst()
	b.languageMatcher = language.NewMatcher(b.languages)
	return b
}

// SupportedLanguages returns all language tags supported by this bundle.
func (b *I18n) SupportedLanguages() []language.Tag {
	return b.languages
}

// ensureDefaultLanguageFirst ensures the default language is the first element
// in the languages slice, adding it if absent or moving it to the front.
func (b *I18n) ensureDefaultLanguageFirst() {
	switch {
	case len(b.languages) == 0:
		b.languages = []language.Tag{b.defaultLanguage}
	case b.languages[0] == b.defaultLanguage:
		return
	default:
		if i := slices.Index(b.languages, b.defaultLanguage); i > 0 {
			b.languages = slices.Delete(b.languages, i, i+1)
		}
		b.languages = slices.Insert(b.languages, 0, b.defaultLanguage)
	}
}

// matchExactLocale returns the string form of the supported locale that
// exactly matches the given locale, or an empty string if none matches.
func (b *I18n) matchExactLocale(locale string) string {
	_, i, conf := b.languageMatcher.Match(language.Make(locale))
	if conf == language.Exact {
		return b.languages[i].String()
	}
	return ""
}

// IsLanguageSupported reports whether the given language tag can be matched
// to a supported locale. Languages not returned by [I18n.SupportedLanguages]
// may still be supported through the bundle's language matcher.
func (b *I18n) IsLanguageSupported(lang language.Tag) bool {
	_, _, conf := b.languageMatcher.Match(lang)
	return conf > language.No
}

// NewLocalizer creates a [Localizer] for the first matching locale from the
// given candidates. If none match, the default locale is used.
func (b *I18n) NewLocalizer(locales ...string) *Localizer {
	selected := b.defaultLocale
	for _, loc := range locales {
		matched := b.matchExactLocale(loc)
		if matched == "" {
			continue
		}
		if _, ok := b.parsedTranslations[matched]; ok {
			selected = matched
			break
		}
	}
	return &Localizer{
		bundle: b,
		locale: selected,
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
// If MessageFormat compilation fails, it returns the translation with the raw
// text as a graceful fallback.
func (b *I18n) parseTranslation(locale, name, text string) (*parsedTranslation, error) {
	pt := &parsedTranslation{
		name:   name,
		locale: locale,
		text:   text,
	}

	base, _ := language.MustParse(locale).Base()

	formatter, err := mf.New(base.String(), b.mfOptions)
	if err != nil {
		return pt, nil //nolint:nilerr // Graceful fallback on compilation error
	}

	compiled, err := formatter.Compile(text)
	if err != nil {
		return pt, nil //nolint:nilerr // Graceful fallback on compilation error
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

// formatFallbacks populates missing translations for each locale by looking up
// the best available fallback from the configured fallback chain.
func (b *I18n) formatFallbacks() {
	for _, defTrans := range b.parsedTranslations[b.defaultLocale] {
		for locale, trans := range b.parsedTranslations {
			if locale == b.defaultLocale {
				continue
			}
			if _, ok := trans[defTrans.name]; ok {
				continue
			}
			if best := b.lookupBestFallback(locale, defTrans.name); best != nil {
				b.parsedTranslations[locale][defTrans.name] = best
			}
		}
	}
}

// lookupBestFallback finds the best fallback translation for a given locale and
// translation name by traversing the fallback chain.
func (b *I18n) lookupBestFallback(locale, name string) *parsedTranslation {
	return b.lookupFallback(locale, name, make(map[string]struct{}))
}

// lookupFallback recursively searches the fallback chain for a translation.
// The visited set prevents infinite recursion from circular fallback configs.
func (b *I18n) lookupFallback(locale, name string, visited map[string]struct{}) *parsedTranslation {
	if _, ok := visited[locale]; ok {
		return nil
	}
	visited[locale] = struct{}{}

	chain, ok := b.fallbacks[locale]
	if !ok {
		return b.parsedTranslations[b.defaultLocale][name]
	}
	for _, fb := range chain {
		if v, ok := b.parsedTranslations[fb][name]; ok {
			return v
		}
		if found := b.lookupFallback(fb, name, visited); found != nil {
			return found
		}
	}
	// All explicit fallbacks exhausted; fall back to the default locale.
	return b.parsedTranslations[b.defaultLocale][name]
}
