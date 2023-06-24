package i18n

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goccy/go-json"
	"github.com/gotnospirit/messageformat"
	"golang.org/x/text/language"
)

// Unmarshaler unmarshals the translation files, can be `json.Unmarshal` or `yaml.Unmarshal`.
type Unmarshaler func(data []byte, v any) error

// I18n is the main internationalization core.
type I18n struct {
	defaultLocale             string
	defaultLanguage           language.Tag
	languages                 []language.Tag
	unmarshaler               Unmarshaler
	languageMatcher           language.Matcher // matcher is a language.Matcher configured for all supported languages.
	fallbacks                 map[string][]string
	parsedTranslations        map[string]map[string]*parsedTranslation
	runtimeParsedTranslations map[string]*parsedTranslation
}

// WithUnmarshaler replaces the default translation file unmarshaler.
func WithUnmarshaler(u Unmarshaler) func(*I18n) {
	return func(bundle *I18n) {
		bundle.unmarshaler = u
	}
}

// WithFallback changes fallback settings.
func WithFallback(f map[string][]string) func(*I18n) {
	return func(bundle *I18n) {
		bundle.fallbacks = f
	}
}

func WithDefaultLocale(locale string) func(*I18n) {
	return func(bundle *I18n) {
		bundle.defaultLanguage = language.Make(locale)
		bundle.defaultLocale = bundle.defaultLanguage.String()
	}
}

func WithLocales(languages ...string) func(*I18n) {
	return func(bundle *I18n) {
		var tags []language.Tag
		for _, lang := range languages {
			tag, err := language.Parse(lang)
			if err == nil && tag != language.Und {
				tags = append(tags, tag)
			}
		}
		bundle.languages = tags
	}
}

// New creates a new internationalization.
func NewBundle(options ...func(*I18n)) *I18n {
	bundle := &I18n{
		languages:                 make([]language.Tag, 0),
		unmarshaler:               json.Unmarshal,
		fallbacks:                 make(map[string][]string),
		runtimeParsedTranslations: make(map[string]*parsedTranslation),
		parsedTranslations:        make(map[string]map[string]*parsedTranslation),
	}
	for _, o := range options {
		o(bundle)
	}
	if bundle.defaultLanguage == language.Und {
		bundle.defaultLanguage = bundle.languages[0]
		bundle.defaultLocale = bundle.defaultLanguage.String()
	}
	if len(bundle.languages) > 0 && bundle.languages[0] != bundle.defaultLanguage {
		for i, t := range bundle.languages {
			if t == bundle.defaultLanguage {
				bundle.languages = append(bundle.languages[:i], bundle.languages[i+1:]...)
				break
			}
		}
		bundle.languages = append([]language.Tag{bundle.defaultLanguage}, bundle.languages...)
	} else if len(bundle.languages) == 0 {
		bundle.languages = append(bundle.languages, bundle.defaultLanguage)
	}
	bundle.languageMatcher = language.NewMatcher(bundle.languages)
	return bundle
}

func (bundle *I18n) SupportedLanguages() []language.Tag {
	return bundle.languages
}

func (bundle *I18n) getExactSupportedLocale(locale string) string {
	_, i, confidence := bundle.languageMatcher.Match(language.Make(locale))

	if confidence == language.Exact {
		return bundle.languages[i].String()
	}

	return ""
}

// IsLanguageSupported indicates whether a language can be translated.
// The check is done by the bundle's matcher and therefore languages that are not returned by
// SupportedLanguages can be supported.
func (bundle *I18n) IsLanguageSupported(lang language.Tag) bool {
	_, _, confidence := bundle.languageMatcher.Match(lang)
	return confidence > language.No
}

// NewLocalizer reads a locale from the internationalization core.
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

var contextRegExp = regexp.MustCompile("<(.*?)>$")

// parsedTranslation
type parsedTranslation struct {
	locale string
	name   string
	text   string
	format *messageformat.MessageFormat
}

// trimContext
func trimContext(v string) string {
	return contextRegExp.ReplaceAllString(v, "")
}

// parseTranslation
func (bundle *I18n) parseTranslation(locale, name, text string) (*parsedTranslation, error) {
	parsedTrans := &parsedTranslation{
		name: name,
	}
	parsedTrans.locale = locale
	parsedTrans.text = text
	base, _ := language.MustParse(locale).Base()

	langParser, err := messageformat.NewWithCulture(base.String())
	if err != nil {
		return nil, err
	}

	parsedTrans.format, err = langParser.Parse(text)
	if err != nil {
		return nil, err
	}

	return parsedTrans, nil
}

// nameInsenstive converts `zh_CN.music.json`, `zh_CN` and `zh-TW` to `zh-CN`.
func nameInsenstive(v string) string {
	v = filepath.Base(v)
	v = strings.Split(v, ".")[0]
	v = strings.ToLower(v)
	v = strings.ReplaceAll(v, "_", "-")
	return v
}

// formatFallbacks
func (bundle *I18n) formatFallbacks() {
	for _, grandTrans := range bundle.parsedTranslations[bundle.defaultLocale] {
		for locale, trans := range bundle.parsedTranslations {
			//
			if locale == bundle.defaultLocale {
				continue
			}
			//
			if _, ok := trans[grandTrans.name]; !ok {
				if bestfit := bundle.lookupBestFallback(locale, grandTrans.name); bestfit != nil {
					bundle.parsedTranslations[locale][grandTrans.name] = bestfit
				}
			}
		}
	}
}

// lookupBestFallback
func (bundle *I18n) lookupBestFallback(locale, name string) *parsedTranslation {
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
		if j := bundle.lookupBestFallback(fallback, name); j != nil {
			return j
		}
	}
	return nil
}
