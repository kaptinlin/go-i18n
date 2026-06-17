package i18n

import (
	"errors"
	"io/fs"
	"path"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFiles(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t,
		WithDefaultLocale("zh-Hans"),
		WithLocales("zh-Hans"),
	)
	require.NoError(t, bundle.LoadFiles("test/zh-Hans.json", "test/zh_Hans.json", "test/zh-Hans.hello.json"))

	localizer := bundle.NewLocalizer("zh-Hans")
	assert.Equal(t, "讯息 A", localizer.Get("message_a"))
	assert.Equal(t, "讯息 B", localizer.Get("message_b"))
	assert.Equal(t, "讯息 C", localizer.Get("message_c"))
}

func TestLoadGlob(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t,
		WithDefaultLocale("zh-Hans"),
		WithLocales("zh-Hans"),
	)
	require.NoError(t, bundle.LoadGlob("test/*.json"))

	localizer := bundle.NewLocalizer("zh-Hans")
	assert.Equal(t, "讯息 A", localizer.Get("message_a"))
	assert.Equal(t, "讯息 B", localizer.Get("message_b"))
	assert.Equal(t, "讯息 C", localizer.Get("message_c"))
}

func TestLoadFS(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t,
		WithDefaultLocale("zh-Hans"),
		WithLocales("zh-Hans"),
	)
	require.NoError(t, bundle.LoadFS(testTranslationFS, "test/*.json"))

	localizer := bundle.NewLocalizer("zh-Hans")
	assert.Equal(t, "讯息 A", localizer.Get("message_a"))
	assert.Equal(t, "讯息 B", localizer.Get("message_b"))
	assert.Equal(t, "讯息 C", localizer.Get("message_c"))
}

func TestLoadFilesReadError(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t,
		WithDefaultLocale("en"),
		WithLocales("en"),
	)

	err := bundle.LoadFiles("nonexistent/file.json")
	require.Error(t, err)

	var pathErr *fs.PathError
	require.ErrorAs(t, err, &pathErr)
	assert.Equal(t, "nonexistent/file.json", pathErr.Path)
}

func TestLoadGlobInvalidPattern(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t,
		WithDefaultLocale("en"),
		WithLocales("en"),
	)

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
	bundle := newTestBundle(t,
		WithDefaultLocale("en"),
		WithLocales("en"),
	)

	err := bundle.LoadFS(fsys, "test/*.json")
	require.Error(t, err)
	assert.ErrorIs(t, err, readErr)
}

func TestLoadFSInvalidGlob(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{}
	bundle := newTestBundle(t,
		WithDefaultLocale("en"),
		WithLocales("en"),
	)

	err := bundle.LoadFS(fsys, "[")
	require.Error(t, err)
	assert.ErrorIs(t, err, path.ErrBadPattern)
}

func TestLoadFSRejectsUnmatchedFileLocale(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"test/de.json": &fstest.MapFile{Data: []byte(`{"hello":"Hallo"}`)},
	}
	bundle := newTestBundle(t,
		WithDefaultLocale("en"),
		WithLocales("en"),
	)

	err := bundle.LoadFS(fsys, "test/*.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `load translations from "test/de.json"`)
	assert.Contains(t, err.Error(), `translation locale "de" is not configured`)

	assert.Equal(t, "hello", bundle.NewLocalizer("en").Get("hello"))
}

func TestWithUnmarshalerNilKeepsDefault(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t,
		WithDefaultLocale("zh-Hans"),
		WithLocales("zh-Hans"),
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
	bundle := newTestBundle(t,
		WithDefaultLocale("en"),
		WithLocales("en"),
		WithUnmarshaler(badUnmarshaler),
	)

	err := bundle.LoadFiles("test/zh-Hans.json")
	require.Error(t, err)
	assert.ErrorIs(t, err, unmarshalErr)
}

func TestLoadMessagesReturnsCompilationError(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t,
		WithDefaultLocale("en"),
		WithLocales("en"),
	)

	err := bundle.LoadMessages(map[string]map[string]string{
		"en": {"broken": "Hello, {name"},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrMessageFormatCompilation)
	assert.Contains(t, err.Error(), `compile translation for locale "en" key "broken"`)
}

func TestLoadMessagesRejectsUnmatchedLocale(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t,
		WithDefaultLocale("en"),
		WithLocales("en"),
	)
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

	bundle := newTestBundle(t,
		WithDefaultLocale("en"),
		WithLocales("en"),
	)
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
