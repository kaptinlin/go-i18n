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

// WithUnmarshaler replaces the default JSON unmarshaler for translation files.
func WithUnmarshaler(u Unmarshaler) Option {
	return func(bundle *I18n) {
		bundle.unmarshaler = u
	}
}

// WithFallback configures locale fallback chains. Each key is a locale, and
// its value is an ordered list of locales to try before the default locale.
func WithFallback(f map[string][]string) Option {
	return func(bundle *I18n) {
		bundle.fallbacks = f
	}
}

// WithDefaultLocale sets the default locale, used as the ultimate fallback
// when no translation is found in the requested or fallback locales.
func WithDefaultLocale(locale string) Option {
	return func(bundle *I18n) {
		bundle.defaultLanguage = language.Make(locale)
		bundle.defaultLocale = bundle.defaultLanguage.String()
	}
}

// WithLocales configures the supported locales for the bundle.
// Invalid locale strings are silently ignored.
func WithLocales(locales ...string) Option {
	return func(bundle *I18n) {
		tags := make([]language.Tag, 0, len(locales))
		for _, loc := range locales {
			tag, err := language.Parse(loc)
			if err == nil && tag != language.Und {
				tags = append(tags, tag)
			}
		}
		bundle.languages = tags
	}
}

// WithMessageFormatOptions sets MessageFormat options for the bundle.
func WithMessageFormatOptions(opts *mf.MessageFormatOptions) Option {
	return func(bundle *I18n) {
		bundle.mfOptions = opts
	}
}

// WithCustomFormatters sets custom formatters for MessageFormat.
// If no MessageFormat options have been set, a new options struct is created.
func WithCustomFormatters(formatters map[string]any) Option {
	return func(bundle *I18n) {
		if bundle.mfOptions == nil {
			bundle.mfOptions = &mf.MessageFormatOptions{}
		}
		bundle.mfOptions.CustomFormatters = formatters
	}
}

// WithStrictMode enables or disables strict parsing mode for MessageFormat.
// If no MessageFormat options have been set, a new options struct is created.
func WithStrictMode(strict bool) Option {
	return func(bundle *I18n) {
		if bundle.mfOptions == nil {
			bundle.mfOptions = &mf.MessageFormatOptions{}
		}
		bundle.mfOptions.Strict = strict
	}
}

// NewBundle creates a new internationalization bundle with the given options.
// If no default locale is set, the first locale from [WithLocales] is used;
// if no locales are configured, English is used as the default.
func NewBundle(options ...Option) *I18n {
	bundle := &I18n{
		unmarshaler:               func(data []byte, v any) error { return json.Unmarshal(data, v) },
		fallbacks:                 make(map[string][]string),
		runtimeParsedTranslations: make(map[string]*parsedTranslation),
		parsedTranslations:        make(map[string]map[string]*parsedTranslation),
	}
	for _, o := range options {
		o(bundle)
	}
	if bundle.defaultLanguage == language.Und {
		if len(bundle.languages) == 0 {
			bundle.defaultLanguage = language.English
		} else {
			bundle.defaultLanguage = bundle.languages[0]
		}
		bundle.defaultLocale = bundle.defaultLanguage.String()
	}
	bundle.ensureDefaultLanguageFirst()
	bundle.languageMatcher = language.NewMatcher(bundle.languages)
	return bundle
}

// SupportedLanguages returns all language tags supported by this bundle.
func (bundle *I18n) SupportedLanguages() []language.Tag {
	return bundle.languages
}

// ensureDefaultLanguageFirst ensures the default language is the first element
// in the languages slice, adding it if absent or moving it to the front.
func (bundle *I18n) ensureDefaultLanguageFirst() {
	if len(bundle.languages) == 0 {
		bundle.languages = []language.Tag{bundle.defaultLanguage}
		return
	}
	if bundle.languages[0] == bundle.defaultLanguage {
		return
	}
	if i := slices.Index(bundle.languages, bundle.defaultLanguage); i > 0 {
		bundle.languages = slices.Delete(bundle.languages, i, i+1)
	}
	bundle.languages = slices.Insert(bundle.languages, 0, bundle.defaultLanguage)
}

func (bundle *I18n) getExactSupportedLocale(locale string) string {
	_, i, confidence := bundle.languageMatcher.Match(language.Make(locale))

	if confidence == language.Exact {
		return bundle.languages[i].String()
	}

	return ""
}

// IsLanguageSupported reports whether the given language tag can be matched
// to a supported locale. Languages not returned by [I18n.SupportedLanguages]
// may still be supported through the bundle's language matcher.
func (bundle *I18n) IsLanguageSupported(lang language.Tag) bool {
	_, _, confidence := bundle.languageMatcher.Match(lang)
	return confidence > language.No
}

// NewLocalizer creates a [Localizer] for the first matching locale from the
// given candidates. If none match, the default locale is used.
func (bundle *I18n) NewLocalizer(locales ...string) *Localizer {
	selectedLocale := bundle.defaultLocale
	for _, locale := range locales {
		locale = bundle.getExactSupportedLocale(locale)
		if locale != "" {
			if _, ok := bundle.parsedTranslations[locale]; ok {
				selectedLocale = locale
				break
			}
		}
	}

	return &Localizer{
		bundle: bundle,
		locale: selectedLocale,
	}
}

// parsedTranslation holds a pre-compiled translation with its locale, name,
// original text, and an optional compiled MessageFormat function.
type parsedTranslation struct {
	locale string
	name   string
	text   string
	format mf.MessageFunction
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
func (bundle *I18n) parseTranslation(locale, name, text string) (*parsedTranslation, error) {
	parsedTrans := &parsedTranslation{
		name:   name,
		locale: locale,
		text:   text,
	}

	base, _ := language.MustParse(locale).Base()

	// Create new MessageFormat instance
	messageFormat, err := mf.New(base.String(), bundle.mfOptions)
	if err != nil {
		return parsedTrans, nil //nolint:nilerr // Intentionally ignore error for graceful fallback
	}

	compiled, err := messageFormat.Compile(text)
	if err != nil {
		return parsedTrans, nil //nolint:nilerr // Intentionally ignore error for graceful fallback
	}

	parsedTrans.format = compiled
	return parsedTrans, nil
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
func (bundle *I18n) formatFallbacks() {
	for _, grandTrans := range bundle.parsedTranslations[bundle.defaultLocale] {
		for locale, trans := range bundle.parsedTranslations {
			if locale == bundle.defaultLocale {
				continue
			}
			if _, ok := trans[grandTrans.name]; !ok {
				if bestfit := bundle.lookupBestFallback(locale, grandTrans.name); bestfit != nil {
					bundle.parsedTranslations[locale][grandTrans.name] = bestfit
				}
			}
		}
	}
}

// lookupBestFallback finds the best fallback translation for a given locale and
// translation name by traversing the fallback chain.
func (bundle *I18n) lookupBestFallback(locale, name string) *parsedTranslation {
	return bundle.lookupFallback(locale, name, make(map[string]struct{}))
}

// lookupFallback recursively searches the fallback chain for a translation.
// The visited set prevents infinite recursion from circular fallback configs.
func (bundle *I18n) lookupFallback(locale, name string, visited map[string]struct{}) *parsedTranslation {
	if _, ok := visited[locale]; ok {
		return nil
	}
	visited[locale] = struct{}{}

	fallbacks, ok := bundle.fallbacks[locale]
	if !ok {
		if v, ok := bundle.parsedTranslations[bundle.defaultLocale][name]; ok {
			return v
		}
	}
	for _, fallback := range fallbacks {
		if v, ok := bundle.parsedTranslations[fallback][name]; ok {
			return v
		}
		if j := bundle.lookupFallback(fallback, name, visited); j != nil {
			return j
		}
	}
	// All explicit fallbacks exhausted; fall back to the default locale.
	if v, ok := bundle.parsedTranslations[bundle.defaultLocale][name]; ok {
		return v
	}
	return nil
}
