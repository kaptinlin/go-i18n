package i18n

import (
	"fmt"

	mf "github.com/kaptinlin/messageformat-go/v1"
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

// Get returns the translation for name with optional MessageFormat variables.
// Returns name as fallback if no translation is found.
func (l *Localizer) Get(name string, data ...Vars) string {
	pt, _ := l.resolve(name)
	return l.localize(pt, data...)
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
	pt, found := l.resolve(name)
	source := TranslationSourceMissing
	if found {
		source = TranslationSourceFallback
		if pt.locale == l.locale {
			source = TranslationSourceDirect
		}
	}
	return TranslationResult{
		Text:   l.localize(pt, data...),
		Locale: pt.locale,
		Source: source,
	}
}

func (l *Localizer) resolve(name string) (*parsedTranslation, bool) {
	if pt, ok := l.bundle.parsedTranslations[l.locale][name]; ok {
		return pt, true
	}
	return l.bundle.getRuntimeParsedTranslation(name), false
}

func (l *Localizer) localize(pt *parsedTranslation, data ...Vars) string {
	params := varsToParams(data)
	if pt.format == nil || params == nil {
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
// This bypasses translation lookup and recompiles the message on each call,
// so it is intended for dynamic, non-hot-path messages that are not stored in
// translation files. Prefer [Localizer.Get] for normal translated content.
func (l *Localizer) Format(message string, data ...Vars) (string, error) {
	base, err := messageFormatBase(l.locale)
	if err != nil {
		return "", err
	}

	formatter, err := mf.New(base, l.bundle.mfOptions)
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

func varsToParams(data []Vars) any {
	if len(data) == 0 {
		return nil
	}
	return map[string]any(data[0])
}
