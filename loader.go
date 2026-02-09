package i18n

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
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

// LoadFiles loads the translations from the given file paths.
func (bundle *I18n) LoadFiles(files ...string) error {
	data := make(map[string]map[string]string)

	for _, file := range files {
		b, err := os.ReadFile(file) //nolint:gosec
		if err != nil {
			return fmt.Errorf("read translation file %q: %w", file, err)
		}
		if err := bundle.mergeTranslation(data, file, b); err != nil {
			return err
		}
	}
	return bundle.LoadMessages(data)
}

// LoadGlob loads the translations from files matching the specified
// glob patterns.
func (bundle *I18n) LoadGlob(patterns ...string) error {
	files, err := collectGlobs(patterns, func(p string) ([]string, error) {
		return filepath.Glob(p)
	})
	if err != nil {
		return err
	}
	return bundle.LoadFiles(files...)
}

// LoadFS loads translations from an [fs.FS], useful for go:embed.
func (bundle *I18n) LoadFS(fsys fs.FS, patterns ...string) error {
	files, err := collectGlobs(patterns, func(p string) ([]string, error) {
		return fs.Glob(fsys, p)
	})
	if err != nil {
		return err
	}

	data := make(map[string]map[string]string)
	for _, file := range files {
		b, err := fs.ReadFile(fsys, file)
		if err != nil {
			return fmt.Errorf("read translation file %q: %w", file, err)
		}
		if err := bundle.mergeTranslation(data, file, b); err != nil {
			return err
		}
	}
	return bundle.LoadMessages(data)
}

// mergeTranslation unmarshals raw file bytes and merges the resulting
// translations into data, keyed by the locale derived from the file name.
func (bundle *I18n) mergeTranslation(data map[string]map[string]string, file string, b []byte) error {
	var trans map[string]string
	if err := bundle.unmarshaler(b, &trans); err != nil {
		return fmt.Errorf("unmarshal translation file %q: %w", file, err)
	}
	locale := nameInsensitive(file)
	if _, ok := data[locale]; !ok {
		data[locale] = make(map[string]string, len(trans))
	}
	maps.Copy(data[locale], trans)
	return nil
}

// collectGlobs expands each pattern using globFn, deduplicates the
// results, and returns them in sorted order.
func collectGlobs(patterns []string, globFn func(string) ([]string, error)) ([]string, error) {
	var files []string
	for _, p := range patterns {
		matches, err := globFn(p)
		if err != nil {
			return nil, fmt.Errorf("glob pattern %q: %w", p, err)
		}
		files = append(files, matches...)
	}
	slices.Sort(files)
	return slices.Compact(files), nil
}
