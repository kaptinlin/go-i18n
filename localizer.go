package i18n

import "fmt"

// Localizer represents a translated locale.
type Localizer struct {
	bundle *I18n

	locale string
}

// Localizer returns the current locale name.
func (localizer *Localizer) Locale() string {
	return localizer.locale
}

// String returns a translated string.
func (localizer *Localizer) Get(name string, data ...Vars) string {
	selectedTrans, err := localizer.lookup(name)
	if err != nil {
		return name
	}

	return localizer.localize(selectedTrans, data...)
}

// GetX returns a translated string with a specified context.
func (localizer *Localizer) GetX(name, context string, data ...Vars) string {
	return localizer.Get(fmt.Sprintf("%s <%s>", name, context), data...)
}

// String returns a translated string with sprintf support.
func (localizer *Localizer) Getf(name string, data ...interface{}) string {
	selectedTrans, err := localizer.lookup(name)
	if err != nil {
		return name
	}

	return fmt.Sprintf(localizer.localize(selectedTrans), data...)
}

// lookup
func (localizer *Localizer) lookup(name string) (*parsedTranslation, error) {
	if selectedTrans, ok := localizer.bundle.parsedTranslations[localizer.locale][name]; ok {
		return selectedTrans, nil
	}
	runtimeTrans, ok := localizer.bundle.runtimeParsedTranslations[name]
	if !ok {
		var err error
		runtimeTrans, err = localizer.bundle.parseTranslation(localizer.bundle.defaultLocale, name, trimContext(name))
		if err != nil {
			return nil, err
		}
	}
	localizer.bundle.runtimeParsedTranslations[name] = runtimeTrans
	return runtimeTrans, nil
}

// localize
func (localizer *Localizer) localize(tran *parsedTranslation, data ...Vars) string {
	if len(data) == 0 {
		return tran.text
	}

	if tran.format != nil {
		str, err := tran.format.FormatMap(data[0])

		if err == nil {
			return str
		}
	}
	return tran.text
}
