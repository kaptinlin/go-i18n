package i18n

import "golang.org/x/text/language"

// MatchAvailableLocale returns the best matching locale from the bundle's
// configured locales for the given Accept-Language header strings. If no
// confident match is found, the default locale is returned.
func (b *I18n) MatchAvailableLocale(accepts ...string) string {
	tags := make([]language.Tag, 0, max(len(accepts)*3, 4))
	for _, s := range accepts {
		parsed, _, err := language.ParseAcceptLanguage(s)
		if err != nil {
			continue
		}
		tags = append(tags, parsed...)
	}

	if len(tags) == 0 {
		return b.languages[0].String()
	}

	_, idx, conf := b.languageMatcher.Match(tags...)
	if conf > language.No {
		return b.languages[idx].String()
	}

	return b.languages[0].String()
}
