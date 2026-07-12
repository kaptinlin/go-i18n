package i18n

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

type stagedMessage struct {
	text string
	path string
}

// LoadMessages loads translations from a locale-keyed map. Within one call,
// locale aliases may contribute disjoint keys, but a duplicate canonical
// locale and key is rejected.
func (i *I18n) LoadMessages(msgs map[string]map[string]string) error {
	staged := make(map[string]map[string]stagedMessage, len(msgs))
	for locale, texts := range msgs {
		staged[locale] = make(map[string]stagedMessage, len(texts))
		for name, text := range texts {
			staged[locale][name] = stagedMessage{text: text}
		}
	}
	return i.loadMessages(staged)
}

func (i *I18n) loadMessages(msgs map[string]map[string]stagedMessage) error {
	i.catalogMu.Lock()
	defer i.catalogMu.Unlock()

	translations := cloneTranslations(i.directTranslations)
	origins := make(map[string]map[string]string, len(msgs))
	for _, loc := range slices.Sorted(maps.Keys(msgs)) {
		messages := msgs[loc]
		locale, err := i.resolveLoadLocale(loc)
		if err != nil {
			return err
		}
		if translations[locale] == nil {
			translations[locale] = make(map[string]*parsedTranslation, len(messages))
		}
		if origins[locale] == nil {
			origins[locale] = make(map[string]string, len(messages))
		}
		for _, name := range slices.Sorted(maps.Keys(messages)) {
			if first, ok := origins[locale][name]; ok {
				return fmt.Errorf(
					"locale %q key %q declared by locale inputs %q and %q",
					locale, name, first, loc,
				)
			}
			origins[locale][name] = loc

			message := messages[name]
			pt, err := i.parseTranslation(locale, name, message.text)
			if err != nil {
				if message.path != "" {
					return fmt.Errorf("load translations from %q: %w", message.path, err)
				}
				return err
			}
			translations[locale][name] = pt
		}
	}
	i.directTranslations = translations
	return nil
}

// LoadFiles loads translations from the given file paths. File errors retain
// the source path through decoding and MessageFormat compilation.
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
	msgs := make(map[string]map[string]stagedMessage, len(files))
	for _, f := range files {
		raw, err := readFn(f)
		if err != nil {
			return fmt.Errorf("read file %q: %w", f, err)
		}
		if err := i.mergeTranslation(msgs, f, raw); err != nil {
			return fmt.Errorf("load translations from %q: %w", f, err)
		}
	}
	return i.loadMessages(msgs)
}

func (i *I18n) mergeTranslation(
	msgs map[string]map[string]stagedMessage, file string, raw []byte,
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
		msgs[locale] = make(map[string]stagedMessage, len(kv))
	}
	for _, key := range slices.Sorted(maps.Keys(kv)) {
		if first, ok := msgs[locale][key]; ok {
			return fmt.Errorf(
				"locale %q key %q declared in %q and %q",
				locale, key, first.path, file,
			)
		}
		msgs[locale][key] = stagedMessage{text: kv[key], path: file}
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
