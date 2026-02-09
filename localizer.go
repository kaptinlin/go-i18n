package i18n

import (
	"fmt"

	mf "github.com/kaptinlin/messageformat-go/v1"
	"golang.org/x/text/language"
)

// Localizer provides translation methods for a specific locale. Create one
// via [I18n.NewLocalizer].
type Localizer struct {
	bundle *I18n
	locale string
}

// Locale returns the resolved locale name for this localizer.
func (l *Localizer) Locale() string {
	return l.locale
}

// Get returns the translation for name, applying optional MessageFormat
// variables. If no translation is found, name itself is returned as a
// fallback.
func (l *Localizer) Get(name string, data ...Vars) string {
	pt, err := l.lookup(name)
	if err != nil {
		return name
	}
	return l.localize(pt, data...)
}

// GetX returns the translation for name disambiguated by context, applying
// optional MessageFormat variables. The context is appended as " <context>"
// to form the lookup key (e.g., "Post <verb>").
func (l *Localizer) GetX(name, context string, data ...Vars) string {
	return l.Get(name+" <"+context+">", data...)
}

// Getf returns the translation for name, then applies [fmt.Sprintf]
// formatting with the provided arguments. If no translation is found,
// name itself is used as the format string.
func (l *Localizer) Getf(name string, args ...any) string {
	pt, err := l.lookup(name)
	if err != nil {
		return name
	}
	return fmt.Sprintf(l.localize(pt), args...)
}

// lookup resolves the translation for name by checking the locale's
// pre-parsed translations first, then falling back to runtime-parsed
// translations from the default locale.
func (l *Localizer) lookup(name string) (*parsedTranslation, error) {
	if pt, ok := l.bundle.parsedTranslations[l.locale][name]; ok {
		return pt, nil
	}
	if pt, ok := l.bundle.runtimeParsedTranslations[name]; ok {
		return pt, nil
	}
	pt, err := l.bundle.parseTranslation(l.bundle.defaultLocale, name, trimContext(name))
	if err != nil {
		return nil, err
	}
	l.bundle.runtimeParsedTranslations[name] = pt
	return pt, nil
}

// localize formats a parsed translation with the given variables.
// Without variables the raw text is returned. With variables and a
// compiled MessageFormat function, the formatted result is returned.
func (l *Localizer) localize(pt *parsedTranslation, data ...Vars) string {
	if len(data) == 0 || pt.format == nil {
		return pt.text
	}
	result, err := pt.format(map[string]any(data[0]))
	if err != nil {
		return pt.text
	}
	if str, ok := result.(string); ok {
		return str
	}
	return pt.text
}

// Format compiles and formats a MessageFormat message directly, bypassing
// the translation lookup. This is useful for formatting dynamic messages
// that are not stored in translation files. Returns the formatted string
// or an error if compilation or formatting fails.
func (l *Localizer) Format(message string, data ...Vars) (string, error) {
	base, _ := language.MustParse(l.locale).Base()

	formatter, err := mf.New(base.String(), l.bundle.mfOptions)
	if err != nil {
		return "", fmt.Errorf("create message format: %w", err)
	}

	compiled, err := formatter.Compile(message)
	if err != nil {
		return "", fmt.Errorf("compile message: %w", err)
	}

	var params any
	if len(data) > 0 {
		params = map[string]any(data[0])
	}

	result, err := compiled(params)
	if err != nil {
		return "", fmt.Errorf("format message: %w", err)
	}

	str, ok := result.(string)
	if !ok {
		return fmt.Sprintf("%v", result), nil
	}
	return str, nil
}
