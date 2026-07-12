package i18n

import (
	"strings"

	"golang.org/x/text/language"
)

// MatchAvailableLocale returns the best matching locale for the given
// Accept-Language header strings. Returns the default locale if no match is found.
func (i *I18n) MatchAvailableLocale(accepts ...string) string {
	locale, ok := i.matchAvailableLocale(accepts...)
	if !ok {
		return i.defaultLocale
	}
	return locale
}

func (i *I18n) matchAvailableLocale(accepts ...string) (string, bool) {
	tags, _, err := language.ParseAcceptLanguage(strings.Join(accepts, ","))
	if err != nil || len(tags) == 0 {
		return "", false
	}

	_, idx, conf := i.languageMatcher.Match(tags...)
	if conf > language.No {
		return i.languages[idx].String(), true
	}

	return "", false
}
