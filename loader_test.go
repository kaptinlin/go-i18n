package i18n

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFiles(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, "zh-Hans")
	require.NoError(t, bundle.LoadFiles("test/zh-Hans.json", "test/zh_Hans.json", "test/zh-Hans.hello.json"))

	localizer := bundle.NewLocalizer("zh-Hans")
	assert.Equal(t, "讯息 A", localizer.Get("message_a"))
	assert.Equal(t, "讯息 B", localizer.Get("message_b"))
	assert.Equal(t, "讯息 C", localizer.Get("message_c"))
}

func TestLoadGlob(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, "zh-Hans")
	require.NoError(t, bundle.LoadGlob("test/*.json"))

	localizer := bundle.NewLocalizer("zh-Hans")
	assert.Equal(t, "讯息 A", localizer.Get("message_a"))
	assert.Equal(t, "讯息 B", localizer.Get("message_b"))
	assert.Equal(t, "讯息 C", localizer.Get("message_c"))
}

func TestLoadFS(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, "zh-Hans")
	require.NoError(t, bundle.LoadFS(testTranslationFS, "test/*.json"))

	localizer := bundle.NewLocalizer("zh-Hans")
	assert.Equal(t, "讯息 A", localizer.Get("message_a"))
	assert.Equal(t, "讯息 B", localizer.Get("message_b"))
	assert.Equal(t, "讯息 C", localizer.Get("message_c"))
}

func TestLoadFSRejectsDuplicateCanonicalLocaleKeyDeclarations(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"locales/zh-Hans.base.json":  &fstest.MapFile{Data: []byte(`{"shared":"first","new":"new"}`)},
		"locales/zh_Hans.extra.json": &fstest.MapFile{Data: []byte(`{"shared":"second"}`)},
	}
	bundle := newTestBundle(t, "zh-Hans")
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"zh-Hans": {"keep": "kept"},
	}))

	err := bundle.LoadFS(fsys, "locales/*.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `locale "zh-Hans" key "shared"`)
	assert.Contains(t, err.Error(), `"locales/zh-Hans.base.json"`)
	assert.Contains(t, err.Error(), `"locales/zh_Hans.extra.json"`)
	assert.Equal(t, "kept", bundle.NewLocalizer("zh-Hans").Get("keep"))
	assert.False(t, bundle.Has("zh-Hans", "new"))
	assert.False(t, bundle.Has("zh-Hans", "shared"))
}

func TestLoadFSAllowsReplacementAcrossCalls(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"locales/en.base.json":  &fstest.MapFile{Data: []byte(`{"shared":"first"}`)},
		"locales/en.extra.json": &fstest.MapFile{Data: []byte(`{"shared":"second"}`)},
	}
	bundle := newTestBundle(t, "en")

	require.NoError(t, bundle.LoadFS(fsys, "locales/en.base.json"))
	require.NoError(t, bundle.LoadFS(fsys, "locales/en.extra.json"))
	assert.Equal(t, "second", bundle.NewLocalizer("en").Get("shared"))
}

func TestLoadMessagesRejectsDuplicateCanonicalLocaleKeyDeclarations(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, "zh-Hans")
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"zh-Hans": {"keep": "kept"},
	}))

	err := bundle.LoadMessages(map[string]map[string]string{
		"ZH_HANS": {
			"new":    "new",
			"shared": "first",
		},
		"zh_Hans": {"shared": "second"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `locale "zh-Hans" key "shared"`)
	assert.Contains(t, err.Error(), `"ZH_HANS"`)
	assert.Contains(t, err.Error(), `"zh_Hans"`)
	assert.Equal(t, "kept", bundle.NewLocalizer("zh-Hans").Get("keep"))
	assert.False(t, bundle.Has("zh-Hans", "new"))
	assert.False(t, bundle.Has("zh-Hans", "shared"))
}

func TestLoadMessagesAllowsDisjointCanonicalLocaleAliases(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, "zh-Hans")
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"zh-Hans": {"first": "first"},
		"zh_Hans": {"second": "second"},
	}))

	localizer := bundle.NewLocalizer("zh-Hans")
	assert.Equal(t, "first", localizer.Get("first"))
	assert.Equal(t, "second", localizer.Get("second"))
}

func TestLoadMessagesAllowsReplacementAcrossCalls(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, "en")
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en": {"shared": "first"},
	}))
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en": {"shared": "second"},
	}))

	assert.Equal(t, "second", bundle.NewLocalizer("en").Get("shared"))
}

func TestLoadFilesReadError(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, "en")

	err := bundle.LoadFiles("nonexistent/file.json")
	require.Error(t, err)

	var pathErr *fs.PathError
	require.ErrorAs(t, err, &pathErr)
	assert.Equal(t, "nonexistent/file.json", pathErr.Path)
}

func TestLoadGlobInvalidPattern(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, "en")

	err := bundle.LoadGlob("[")
	require.Error(t, err)
	assert.ErrorIs(t, err, filepath.ErrBadPattern)
}

type brokenReadFS struct {
	fstest.MapFS
	err error
}

func (f brokenReadFS) ReadFile(name string) ([]byte, error) {
	if name == "test/zh-Hans.json" {
		return nil, f.err
	}
	return fs.ReadFile(f.MapFS, name)
}

func TestLoadFSReadError(t *testing.T) {
	t.Parallel()

	readErr := errors.New("read failed")
	fsys := brokenReadFS{
		MapFS: fstest.MapFS{
			"test/zh-Hans.json": &fstest.MapFile{Data: []byte(`{"message_a":"讯息 A"}`)},
		},
		err: readErr,
	}
	bundle := newTestBundle(t, "en")

	err := bundle.LoadFS(fsys, "test/*.json")
	require.Error(t, err)
	assert.ErrorIs(t, err, readErr)
}

func TestLoadFSInvalidGlob(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{}
	bundle := newTestBundle(t, "en")

	err := bundle.LoadFS(fsys, "[")
	require.Error(t, err)
	assert.ErrorIs(t, err, path.ErrBadPattern)
}

func TestLoadFSNilFilesystemReturnsError(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, "en")

	var err error
	require.NotPanics(t, func() {
		err = bundle.LoadFS(nil, "*.json")
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, fs.ErrInvalid)
	assert.Contains(t, err.Error(), "load translations from filesystem")
}

func TestLoadFSRejectsUnmatchedFileLocale(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"test/de.json": &fstest.MapFile{Data: []byte(`{"hello":"Hallo"}`)},
	}
	bundle := newTestBundle(t, "en")

	err := bundle.LoadFS(fsys, "test/*.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `load translations from "test/de.json"`)
	assert.Contains(t, err.Error(), `translation locale "de" is not configured`)

	assert.Equal(t, "hello", bundle.NewLocalizer("en").Get("hello"))
}

func TestWithUnmarshalerNilKeepsDefault(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, "zh-Hans",
		WithUnmarshaler(nil),
	)

	require.NoError(t, bundle.LoadFiles("test/zh-Hans.json"))

	localizer := bundle.NewLocalizer("zh-Hans")
	assert.Equal(t, "讯息 B", localizer.Get("message_b"))
}

func TestMergeTranslationUnmarshalError(t *testing.T) {
	t.Parallel()

	unmarshalErr := errors.New("unmarshal failed")
	badUnmarshaler := func([]byte, any) error {
		return unmarshalErr
	}
	bundle := newTestBundle(t, "en",
		WithUnmarshaler(badUnmarshaler),
	)

	err := bundle.LoadFiles("test/zh-Hans.json")
	require.Error(t, err)
	assert.ErrorIs(t, err, unmarshalErr)
}

func TestLoadMessagesReturnsCompilationError(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, "en")

	err := bundle.LoadMessages(map[string]map[string]string{
		"en": {"broken": "Hello, {name"},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrMessageFormatCompilation)
	assert.Contains(t, err.Error(), `compile translation for locale "en" key "broken"`)
	assert.NotContains(t, err.Error(), "load translations from")
}

func TestFileLoaderCompilationErrorsIdentifySource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		load func(*I18n, string) error
	}{
		{name: "files", load: func(bundle *I18n, file string) error {
			return bundle.LoadFiles(file)
		}},
		{name: "glob", load: func(bundle *I18n, file string) error {
			return bundle.LoadGlob(filepath.Join(filepath.Dir(file), "*.json"))
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			file := filepath.Join(t.TempDir(), "en.json")
			require.NoError(t, os.WriteFile(file, []byte(`{"broken":"Hello, {name"}`), 0o600))
			bundle := newTestBundle(t, "en")

			err := tt.load(bundle, file)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrMessageFormatCompilation)
			assert.Contains(t, err.Error(), file)
			assert.Contains(t, err.Error(), `locale "en" key "broken"`)
		})
	}
}

func TestLoadFSCompilationErrorIdentifiesSourceAndRollsBack(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"locales/en.bad.json":  &fstest.MapFile{Data: []byte(`{"broken":"Hello, {name"}`)},
		"locales/en.good.json": &fstest.MapFile{Data: []byte(`{"new":"New"}`)},
	}
	bundle := newTestBundle(t, "en")
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en": {"keep": "Old"},
	}))

	err := bundle.LoadFS(fsys, "locales/*.json")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrMessageFormatCompilation)
	assert.Contains(t, err.Error(), `"locales/en.bad.json"`)
	assert.NotContains(t, err.Error(), `"locales/en.good.json"`)
	assert.Contains(t, err.Error(), `locale "en" key "broken"`)

	assert.Equal(t, "Old", bundle.NewLocalizer("en").Get("keep"))
	assert.False(t, bundle.Has("en", "new"))
	assert.False(t, bundle.Has("en", "broken"))
}

func TestLoadMessagesRejectsUnmatchedLocale(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, "en")
	err := bundle.LoadMessages(map[string]map[string]string{
		"en": {"hello": "Hello"},
		"de": {"hello": "Hallo"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `translation locale "de" is not configured`)

	loc := bundle.NewLocalizer("en")
	assert.Equal(t, "hello", loc.Get("hello"))
}

func TestLoadMessagesFailedCompileLeavesCatalogUntouched(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, "en")
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en": {"keep": "Old"},
	}))

	err := bundle.LoadMessages(map[string]map[string]string{
		"en": {
			"a_valid":  "New",
			"z_broken": "Hello, {name",
		},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrMessageFormatCompilation)

	loc := bundle.NewLocalizer("en")
	assert.Equal(t, "Old", loc.Get("keep"))
	assert.Equal(t, "a_valid", loc.Get("a_valid"))
}
