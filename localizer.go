package i18n

import "fmt"

// Localizer provides translation methods for a specific locale. Create one
// via [I18n.NewLocalizer].
type Localizer struct {
	bundle *I18n
	locale string
}

type resolvedTranslation struct {
	translation   *parsedTranslation
	source        TranslationSource
	matchedLocale string
	catalogLocale string
}

// Locale returns the resolved locale name for this localizer.
func (l *Localizer) Locale() string {
	return l.locale
}

// Get returns the translation for name with optional MessageFormat variables.
// Returns name as fallback if no translation is found.
func (l *Localizer) Get(name string, data ...Vars) string {
	result, _ := l.lookup(name, data...)
	return result.Text
}

// GetX returns the translation for name disambiguated by context.
// The context is appended as " <context>" to form the lookup key.
// For example, GetX("Post", "verb") looks up "Post <verb>".
func (l *Localizer) GetX(name, context string, data ...Vars) string {
	return l.Get(name+" <"+context+">", data...)
}

// GetTemplate returns the resolved loaded template for name without formatting.
// The returned template is the raw MessageFormat text from the direct or
// fallback catalog entry that would supply [Localizer.Get]. If no loaded
// translation is found, it returns "", false.
func (l *Localizer) GetTemplate(name string) (string, bool) {
	resolved := l.resolve(name)
	if resolved.source == TranslationSourceMissing {
		return "", false
	}
	return resolved.translation.text, true
}

// Lookup returns the translation for name with full lookup details. If a
// loaded translation cannot be formatted, the result contains its raw template
// and provenance alongside the formatting error. A missing translation is not
// an error. Use [Localizer.Get] for the forgiving string-only path.
func (l *Localizer) Lookup(name string, data ...Vars) (TranslationResult, error) {
	return l.lookup(name, data...)
}

func (l *Localizer) lookup(name string, data ...Vars) (TranslationResult, error) {
	resolved := l.resolve(name)
	text, err := l.localize(resolved.translation, data...)
	result := TranslationResult{
		Text:          text,
		MatchedLocale: resolved.matchedLocale,
		CatalogLocale: resolved.catalogLocale,
		Source:        resolved.source,
	}
	if resolved.source != TranslationSourceMissing {
		result.Template = resolved.translation.text
	}
	if resolved.source == TranslationSourceMissing {
		return result, nil
	}
	if err != nil {
		err = fmt.Errorf(
			"format translation for locale %q key %q: %w",
			resolved.catalogLocale, name, err,
		)
	}
	return result, err
}

func (l *Localizer) resolve(name string) resolvedTranslation {
	translations := l.bundle.catalogSnapshot()
	if pt, ok := translations[l.locale][name]; ok {
		return resolvedTranslation{
			translation:   pt,
			source:        TranslationSourceDirect,
			matchedLocale: l.locale,
			catalogLocale: pt.locale,
		}
	}
	if pt := l.bundle.lookupFallback(translations, l.locale, name); pt != nil {
		return resolvedTranslation{
			translation:   pt,
			source:        TranslationSourceFallback,
			matchedLocale: l.locale,
			catalogLocale: pt.locale,
		}
	}
	return resolvedTranslation{
		translation:   l.bundle.runtimeFallbackTranslation(name),
		source:        TranslationSourceMissing,
		matchedLocale: l.locale,
	}
}

func (l *Localizer) localize(pt *parsedTranslation, data ...Vars) (string, error) {
	if pt.format == nil {
		return pt.text, nil
	}

	result, err := formatCompiled(pt.format, data)
	if err != nil {
		return pt.text, err
	}
	return result, nil
}

// Format compiles and formats a MessageFormat message directly.
// This bypasses translation lookup and recompiles the message on each call,
// so it is intended for dynamic, non-hot-path messages that are not stored in
// translation files. Prefer [Localizer.Get] for normal translated content.
func (l *Localizer) Format(message string, data ...Vars) (string, error) {
	return l.bundle.messageFormat.format(l.locale, message, data)
}
