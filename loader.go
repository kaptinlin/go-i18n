package i18n

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

// LoadMessages populates the bundle with translations from the given
// locale-keyed map. Locales that do not match any configured locale are
// silently skipped.
func (b *I18n) LoadMessages(msgs map[string]map[string]string) error {
	for loc, texts := range msgs {
		locale := b.matchExactLocale(loc)
		if locale == "" {
			continue
		}
		if _, ok := b.parsedTranslations[locale]; !ok {
			b.parsedTranslations[locale] = make(map[string]*parsedTranslation)
		}
		for name, text := range texts {
			pt, err := b.parseTranslation(locale, name, text)
			if err != nil {
				return err
			}
			b.parsedTranslations[locale][name] = pt
		}
	}
	b.formatFallbacks()
	return nil
}

// LoadFiles loads translations from the given file paths.
func (b *I18n) LoadFiles(files ...string) error {
	return b.loadFiles(files, func(name string) ([]byte, error) {
		return os.ReadFile(name) //nolint:gosec
	})
}

// LoadGlob loads translations from files matching the given glob patterns.
func (b *I18n) LoadGlob(patterns ...string) error {
	files, err := collectGlobs(patterns, func(p string) ([]string, error) {
		return filepath.Glob(p)
	})
	if err != nil {
		return err
	}
	return b.LoadFiles(files...)
}

// LoadFS loads translations from an [fs.FS], useful for go:embed.
func (b *I18n) LoadFS(fsys fs.FS, patterns ...string) error {
	files, err := collectGlobs(patterns, func(p string) ([]string, error) {
		return fs.Glob(fsys, p)
	})
	if err != nil {
		return err
	}
	return b.loadFiles(files, func(name string) ([]byte, error) {
		return fs.ReadFile(fsys, name)
	})
}

// loadFiles reads each file using readFn, unmarshals the contents,
// and loads the resulting translations into the bundle.
func (b *I18n) loadFiles(files []string, readFn func(string) ([]byte, error)) error {
	msgs := make(map[string]map[string]string, len(files))
	for _, f := range files {
		raw, err := readFn(f)
		if err != nil {
			return fmt.Errorf("read file %q: %w", f, err)
		}
		if err := b.mergeTranslation(msgs, f, raw); err != nil {
			return err
		}
	}
	return b.LoadMessages(msgs)
}

// mergeTranslation unmarshals raw bytes from file and merges the
// resulting key-value pairs into msgs, keyed by the locale derived
// from the file name.
func (b *I18n) mergeTranslation(
	msgs map[string]map[string]string, file string, raw []byte,
) error {
	var kv map[string]string
	if err := b.unmarshaler(raw, &kv); err != nil {
		return fmt.Errorf("unmarshal %q: %w", file, err)
	}
	locale := nameInsensitive(file)
	if _, ok := msgs[locale]; !ok {
		msgs[locale] = make(map[string]string, len(kv))
	}
	maps.Copy(msgs[locale], kv)
	return nil
}

// collectGlobs expands each pattern using globFn, deduplicates the
// results, and returns them in sorted order.
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
