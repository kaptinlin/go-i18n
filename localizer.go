package i18n

import (
	"errors"
	"fmt"

	mf "github.com/kaptinlin/messageformat-go/v1"
	"golang.org/x/text/language"
)

// ErrTranslationNotFound indicates that a translation was not found in any
// of the loaded translations for the requested locale or its fallback chain.
// A runtime fallback translation is created using the key itself as the text.
var ErrTranslationNotFound = errors.New("translation not found")

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

// Get returns the translation for name with optional MessageFormat variables.
// Returns name as fallback if no translation is found.
func (l *Localizer) Get(name string, data ...Vars) string {
	pt, _ := l.lookup(name)
	return l.localize(pt, data...)
}

// GetX returns the translation for name disambiguated by context.
// The context is appended as " <context>" to form the lookup key.
// For example, GetX("Post", "verb") looks up "Post <verb>".
func (l *Localizer) GetX(name, context string, data ...Vars) string {
	return l.Get(name+" <"+context+">", data...)
}

// GetWithLocale returns the translation for name with optional MessageFormat variables,
// along with the source locale where the translation was found.
// If the translation is not found in the loaded translations, it returns [ErrTranslationNotFound]
// and the source locale will be the default locale where the runtime fallback was created.
func (l *Localizer) GetWithLocale(name string, data ...Vars) (text string, sourceLocale string, err error) {
	pt, err := l.lookup(name)
	return l.localize(pt, data...), pt.locale, err
}

// GetXWithLocale returns the translation for name disambiguated by context,
// along with the source locale where the translation was found.
// If the translation is not found, it returns [ErrTranslationNotFound].
// The context is appended as " <context>" to form the lookup key.
// For example, GetXWithLocale("Post", "verb") looks up "Post <verb>".
func (l *Localizer) GetXWithLocale(name, context string, data ...Vars) (text string, sourceLocale string, err error) {
	return l.GetWithLocale(name+" <"+context+">", data...)
}

// Getf returns the translation for name formatted with fmt.Sprintf.
// Uses name as the format string if no translation is found.
func (l *Localizer) Getf(name string, args ...any) string {
	pt, _ := l.lookup(name)
	return fmt.Sprintf(l.localize(pt), args...)
}

// GetfWithLocale returns the translation for name formatted with fmt.Sprintf,
// along with the source locale where the translation was found.
// If the translation is not found, it returns [ErrTranslationNotFound].
// Note: If the translation is not found, the returned text will be the key itself
// (not formatted) to avoid fmt.Sprintf errors with missing arguments.
func (l *Localizer) GetfWithLocale(name string, args ...any) (text string, sourceLocale string, err error) {
	pt, err := l.lookup(name)
	if err != nil {
		// Return the key as-is without formatting to avoid %!EXTRA errors
		return name, pt.locale, err
	}
	return fmt.Sprintf(l.localize(pt), args...), pt.locale, nil
}

// lookup resolves the translation for name by checking the locale's
// pre-parsed translations first, then falling back to runtime-parsed
// translations from the default locale. If no translation exists, it
// creates a new runtime translation using the name as the text.
//
// Returns [ErrTranslationNotFound] if the translation is not found in loaded
// translations (runtime fallback was used). May also wrap [ErrMessageFormatCompilation]
// if MessageFormat compilation failed during runtime fallback creation.
func (l *Localizer) lookup(name string) (*parsedTranslation, error) {
	if pt, ok := l.bundle.parsedTranslations[l.locale][name]; ok {
		return pt, nil
	}
	if pt, ok := l.bundle.runtimeParsedTranslations[name]; ok {
		return pt, ErrTranslationNotFound
	}
	pt, err := l.bundle.parseTranslation(l.bundle.defaultLocale, name, trimContext(name))
	l.bundle.runtimeParsedTranslations[name] = pt
	if err != nil {
		// MessageFormat compilation failed, but we still have the raw text
		return pt, fmt.Errorf("%w: %w", ErrTranslationNotFound, err)
	}
	return pt, ErrTranslationNotFound
}

// localize formats a parsed translation with the given variables.
// Without variables the raw text is returned. With variables and a
// compiled MessageFormat function, the formatted result is returned.
func (l *Localizer) localize(pt *parsedTranslation, data ...Vars) string {
	if pt.format == nil {
		return pt.text
	}
	params := varsToParams(data)
	if params == nil {
		return pt.text
	}
	result, err := pt.format(params)
	if err != nil {
		return pt.text
	}
	str, ok := result.(string)
	if !ok {
		return pt.text
	}
	return str
}

// Format compiles and formats a MessageFormat message directly.
// This bypasses translation lookup and is useful for dynamic messages
// not stored in translation files.
func (l *Localizer) Format(message string, data ...Vars) (string, error) {
	base, _ := language.MustParse(l.locale).Base()

	formatter, err := mf.New(base.String(), l.bundle.mfOptions)
	if err != nil {
		return "", fmt.Errorf("create formatter: %w", err)
	}

	compiled, err := formatter.Compile(message)
	if err != nil {
		return "", fmt.Errorf("compile message: %w", err)
	}

	params := varsToParams(data)

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

// varsToParams converts optional Vars arguments to a params value
// suitable for a compiled MessageFormat function. Returns nil when
// no variables are provided. Only the first Vars argument is used.
func varsToParams(data []Vars) any {
	if len(data) == 0 {
		return nil
	}
	return map[string]any(data[0])
}
