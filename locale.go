package i18n

import "golang.org/x/text/language"

// MatchAvailableLocale returns the best matching locale from the bundle's
// supported locales for the given Accept-Language header values. If no
// match is found, the default locale is returned.
func (b *I18n) MatchAvailableLocale(locales ...string) string {
	// Estimate capacity: most Accept-Language headers contain 2-4 tags.
	tags := make([]language.Tag, 0, max(len(locales)*3, 4))

	for _, accept := range locales {
		desired, _, err := language.ParseAcceptLanguage(accept)
		if err != nil {
			continue
		}
		tags = append(tags, desired...)
	}

	if _, index, conf := b.languageMatcher.Match(tags...); conf > language.No {
		return b.languages[index].String()
	}

	return b.languages[0].String()
}
