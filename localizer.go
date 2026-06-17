package i18n

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
	resolved := l.resolve(name)
	return l.localize(resolved.translation, data...)
}

// GetX returns the translation for name disambiguated by context.
// The context is appended as " <context>" to form the lookup key.
// For example, GetX("Post", "verb") looks up "Post <verb>".
func (l *Localizer) GetX(name, context string, data ...Vars) string {
	return l.Get(name+" <"+context+">", data...)
}

// Lookup returns the translation for name with full lookup details.
// Use [Localizer.Get] for the common case where only the text is needed.
func (l *Localizer) Lookup(name string, data ...Vars) TranslationResult {
	resolved := l.resolve(name)
	return TranslationResult{
		Text:          l.localize(resolved.translation, data...),
		MatchedLocale: resolved.matchedLocale,
		CatalogLocale: resolved.catalogLocale,
		Source:        resolved.source,
	}
}

func (l *Localizer) resolve(name string) resolvedTranslation {
	if pt, ok := l.bundle.directTranslations[l.locale][name]; ok {
		return resolvedTranslation{
			translation:   pt,
			source:        TranslationSourceDirect,
			matchedLocale: l.locale,
			catalogLocale: pt.locale,
		}
	}
	if pt := l.bundle.lookupFallback(l.locale, name); pt != nil {
		return resolvedTranslation{
			translation:   pt,
			source:        TranslationSourceFallback,
			matchedLocale: l.locale,
			catalogLocale: pt.locale,
		}
	}
	return resolvedTranslation{
		translation:   l.bundle.getRuntimeParsedTranslation(name),
		source:        TranslationSourceMissing,
		matchedLocale: l.locale,
	}
}

func (l *Localizer) localize(pt *parsedTranslation, data ...Vars) string {
	if pt.format == nil {
		return pt.text
	}

	result, err := formatCompiled(pt.format, data)
	if err != nil {
		return pt.text
	}
	return result
}

// Format compiles and formats a MessageFormat message directly.
// This bypasses translation lookup and recompiles the message on each call,
// so it is intended for dynamic, non-hot-path messages that are not stored in
// translation files. Prefer [Localizer.Get] for normal translated content.
func (l *Localizer) Format(message string, data ...Vars) (string, error) {
	return l.bundle.messageFormat.format(l.locale, message, data)
}
