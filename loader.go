package i18n

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

// LoadMessages loads translations from a locale-keyed map. Within one call,
// locale aliases may contribute disjoint keys, but a duplicate canonical
// locale and key is rejected.
func (i *I18n) LoadMessages(msgs map[string]map[string]string) error {
	i.catalogMu.Lock()
	defer i.catalogMu.Unlock()

	translations := cloneTranslations(i.directTranslations)
	origins := make(map[string]map[string]string, len(msgs))
	for _, loc := range slices.Sorted(maps.Keys(msgs)) {
		texts := msgs[loc]
		locale, err := i.resolveLoadLocale(loc)
		if err != nil {
			return err
		}
		if translations[locale] == nil {
			translations[locale] = make(map[string]*parsedTranslation, len(texts))
		}
		if origins[locale] == nil {
			origins[locale] = make(map[string]string, len(texts))
		}
		for _, name := range slices.Sorted(maps.Keys(texts)) {
			if first, ok := origins[locale][name]; ok {
				return fmt.Errorf(
					"locale %q key %q declared by locale inputs %q and %q",
					locale, name, first, loc,
				)
			}
			origins[locale][name] = loc

			text := texts[name]
			pt, err := i.parseTranslation(locale, name, text)
			if err != nil {
				return err
			}
			translations[locale][name] = pt
		}
	}
	i.directTranslations = translations
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
	if fsys == nil {
		return fmt.Errorf("load translations from filesystem: %w", fs.ErrInvalid)
	}

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
	origins := make(map[string]map[string]string, len(files))
	for _, f := range files {
		raw, err := readFn(f)
		if err != nil {
			return fmt.Errorf("read file %q: %w", f, err)
		}
		if err := i.mergeTranslation(msgs, origins, f, raw); err != nil {
			return fmt.Errorf("load translations from %q: %w", f, err)
		}
	}
	return i.LoadMessages(msgs)
}

func (i *I18n) mergeTranslation(
	msgs, origins map[string]map[string]string, file string, raw []byte,
) error {
	var kv map[string]string
	if err := i.unmarshaler(raw, &kv); err != nil {
		return fmt.Errorf("unmarshal %q: %w", file, err)
	}
	locale, err := i.resolveLoadLocale(nameInsensitive(file))
	if err != nil {
		return err
	}
	if msgs[locale] == nil {
		msgs[locale] = make(map[string]string, len(kv))
		origins[locale] = make(map[string]string, len(kv))
	}
	for _, key := range slices.Sorted(maps.Keys(kv)) {
		if first, ok := origins[locale][key]; ok {
			return fmt.Errorf(
				"locale %q key %q declared in %q and %q",
				locale, key, first, file,
			)
		}
		origins[locale][key] = file
		msgs[locale][key] = kv[key]
	}
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

func cloneTranslations(
	translations map[string]map[string]*parsedTranslation,
) map[string]map[string]*parsedTranslation {
	cloned := make(map[string]map[string]*parsedTranslation, len(translations))
	for locale, direct := range translations {
		cloned[locale] = maps.Clone(direct)
	}
	return cloned
}
