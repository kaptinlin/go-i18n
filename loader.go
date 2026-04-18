package i18n

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

// LoadMessages loads translations from a locale-keyed map.
// Locales not matching any configured locale are silently skipped.
func (i *I18n) LoadMessages(msgs map[string]map[string]string) error {
	for loc, texts := range msgs {
		locale := i.matchExactLocale(loc)
		if locale == "" {
			continue
		}
		if i.directTranslations[locale] == nil {
			i.directTranslations[locale] = make(map[string]*parsedTranslation, len(texts))
		}
		if i.parsedTranslations[locale] == nil {
			i.parsedTranslations[locale] = make(map[string]*parsedTranslation, len(texts))
		}
		for name, text := range texts {
			pt, err := i.parseTranslation(locale, name, text)
			if err != nil {
				return err
			}
			i.directTranslations[locale][name] = pt
			i.parsedTranslations[locale][name] = pt
		}
	}
	i.formatFallbacks()
	return nil
}

// LoadFiles loads translations from the given file paths.
func (i *I18n) LoadFiles(files ...string) error {
	return i.loadFiles(files, os.ReadFile)
}

// LoadGlob loads translations from files matching the given glob patterns.
func (i *I18n) LoadGlob(patterns ...string) error {
	files, err := collectGlobs(patterns, filepath.Glob)
	if err != nil {
		return err
	}
	return i.LoadFiles(files...)
}

// LoadFS loads translations from an fs.FS, useful for go:embed.
func (i *I18n) LoadFS(fsys fs.FS, patterns ...string) error {
	files, err := collectGlobs(patterns, func(p string) ([]string, error) {
		return fs.Glob(fsys, p)
	})
	if err != nil {
		return err
	}
	return i.loadFiles(files, func(name string) ([]byte, error) {
		return fs.ReadFile(fsys, name)
	})
}

func (i *I18n) loadFiles(files []string, readFn func(string) ([]byte, error)) error {
	msgs := make(map[string]map[string]string, len(files))
	for _, f := range files {
		raw, err := readFn(f)
		if err != nil {
			return fmt.Errorf("read file %q: %w", f, err)
		}
		if err := i.mergeTranslation(msgs, f, raw); err != nil {
			return fmt.Errorf("load translations from %q: %w", f, err)
		}
	}
	return i.LoadMessages(msgs)
}

func (i *I18n) mergeTranslation(
	msgs map[string]map[string]string, file string, raw []byte,
) error {
	var kv map[string]string
	if err := i.unmarshaler(raw, &kv); err != nil {
		return fmt.Errorf("unmarshal %q: %w", file, err)
	}
	locale := nameInsensitive(file)
	if msgs[locale] == nil {
		msgs[locale] = make(map[string]string, len(kv))
	}
	maps.Copy(msgs[locale], kv)
	return nil
}

func collectGlobs(
	patterns []string, globFn func(string) ([]string, error),
) ([]string, error) {
	paths := make([]string, 0, len(patterns)*4)
	for _, p := range patterns {
		matches, err := globFn(p)
		if err != nil {
			return nil, fmt.Errorf("expand glob %q: %w", p, err)
		}
		paths = append(paths, matches...)
	}
	slices.Sort(paths)
	return slices.Compact(paths), nil
}
