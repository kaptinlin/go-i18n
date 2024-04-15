package i18n

import (
	"io/fs"
	"os"
	"path/filepath"
)

// LoadMessages loads the translations from the map.
func (bundle *I18n) LoadMessages(languages map[string]map[string]string) error {
	for locale, translations := range languages {
		locale = bundle.getExactSupportedLocale(locale)

		if locale != "" {
			if _, ok := bundle.parsedTranslations[locale]; !ok {
				bundle.parsedTranslations[locale] = make(map[string]*parsedTranslation)
			}

			for name, text := range translations {
				trans, err := bundle.parseTranslation(locale, name, text)
				if err != nil {
					return err
				}
				bundle.parsedTranslations[locale][name] = trans
			}
		}
	}
	bundle.formatFallbacks()
	return nil
}

// LoadFiles loads the translations from the files.
func (bundle *I18n) LoadFiles(files ...string) error {
	data := make(map[string]map[string]string)

	for _, file := range files {
		b, err := os.ReadFile(file) //nolint:gosec
		if err != nil {
			return err
		}
		var trans map[string]string
		if err := bundle.unmarshaler(b, &trans); err != nil {
			return err
		}
		locale := nameInsenstive(file)
		_, ok := data[locale]
		if !ok {
			data[locale] = make(map[string]string)
		}
		for name, text := range trans {
			data[locale][name] = text
		}
	}
	return bundle.LoadMessages(data)
}

// LoadGlob loads the translations from the files that matches specified patterns.
func (bundle *I18n) LoadGlob(pattern ...string) error {
	var files []string

	for _, pattern := range pattern {
		v, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}
		files = append(files, v...)
	}

	return bundle.LoadFiles(files...)
}

// LoadFS loads the translation from a `fs.FS`, useful for `go:embed`.
func (bundle *I18n) LoadFS(fsys fs.FS, patterns ...string) error {
	var files []string
	data := make(map[string]map[string]string)

	for _, pattern := range patterns {
		v, err := fs.Glob(fsys, pattern)
		if err != nil {
			return err
		}
		files = append(files, v...)
	}

	for _, file := range files {
		b, err := fs.ReadFile(fsys, file)
		if err != nil {
			return err
		}
		trans := make(map[string]string)
		if err := bundle.unmarshaler(b, &trans); err != nil {
			return err
		}
		locale := nameInsenstive(file)

		_, ok := data[locale]
		if !ok {
			data[locale] = make(map[string]string)
		}
		for name, text := range trans {
			data[locale][name] = text
		}
	}
	return bundle.LoadMessages(data)
}
