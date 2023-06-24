package i18n

import "golang.org/x/text/language"

// MatchAvailableLocale return one of the available locales
func (bundle *I18n) MatchAvailableLocale(locales ...string) string {
	var tags []language.Tag

	for _, accept := range locales {
		desired, _, err := language.ParseAcceptLanguage(accept)
		if err != nil {
			continue
		}
		tags = append(tags, desired...)
	}

	if _, index, conf := bundle.languageMatcher.Match(tags...); conf > language.No {
		return bundle.languages[index].String()
	}

	return bundle.languages[0].String()
}
