package i18n

import (
	"errors"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestLoadFiles(t *testing.T) {
	assert := assert.New(t)

	bundle := NewBundle(
		WithDefaultLocale("zh-Hans"),
		WithLocales("zh-Hans"),
	)
	assert.NoError(bundle.LoadFiles("test/zh-Hans.json", "test/zh_Hans.json", "test/zh-Hans.hello.json"))

	localizer := bundle.NewLocalizer("zh-Hans")
	assert.Equal("讯息 A", localizer.Get("message_a"))
	assert.Equal("讯息 B", localizer.Get("message_b"))
	assert.Equal("讯息 C", localizer.Get("message_c"))
}

func TestLoadGlob(t *testing.T) {
	assert := assert.New(t)

	bundle := NewBundle(
		WithDefaultLocale("zh-Hans"),
		WithLocales("zh-Hans"),
	)
	assert.NoError(bundle.LoadGlob("test/*.json"))

	localizer := bundle.NewLocalizer("zh-Hans")
	assert.Equal("讯息 A", localizer.Get("message_a"))
	assert.Equal("讯息 B", localizer.Get("message_b"))
	assert.Equal("讯息 C", localizer.Get("message_c"))
}

func TestLoadFS(t *testing.T) {
	assert := assert.New(t)

	bundle := NewBundle(
		WithDefaultLocale("zh-Hans"),
		WithLocales("zh-Hans"),
	)
	assert.NoError(bundle.LoadFS(testTranslationFS, "test/*.json"))

	localizer := bundle.NewLocalizer("zh-Hans")
	assert.Equal("讯息 A", localizer.Get("message_a"))
	assert.Equal("讯息 B", localizer.Get("message_b"))
	assert.Equal("讯息 C", localizer.Get("message_c"))
}

func TestLoadFilesReadError(t *testing.T) {
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en"),
	)
	err := bundle.LoadFiles("nonexistent/file.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading translation file")
}

func TestLoadGlobInvalidPattern(t *testing.T) {
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en"),
	)
	// "[" is an invalid glob pattern.
	err := bundle.LoadGlob("[")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expanding glob")
}

func TestLoadFSReadError(t *testing.T) {
	// Use an empty FS so the file doesn't exist.
	fsys := fstest.MapFS{}
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en"),
	)
	err := bundle.LoadFS(fsys, "nonexistent/*.json")
	// No matches means no error (empty load).
	assert.NoError(t, err)
}

func TestLoadFSInvalidGlob(t *testing.T) {
	fsys := fstest.MapFS{}
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en"),
	)
	err := bundle.LoadFS(fsys, "[")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expanding glob")
}

func TestMergeTranslationUnmarshalError(t *testing.T) {
	badUnmarshaler := func([]byte, any) error {
		return errors.New("unmarshal failed")
	}
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en"),
		WithUnmarshaler(badUnmarshaler),
	)
	err := bundle.LoadFiles("test/zh-Hans.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshaling translation file")
}

func TestLoadMessagesSkipsUnmatchedLocale(t *testing.T) {
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en"),
	)
	// "xx" doesn't match any configured locale; should be skipped.
	err := bundle.LoadMessages(map[string]map[string]string{
		"en": {"hello": "Hello"},
		"xx": {"hello": "XX Hello"},
	})
	assert.NoError(t, err)

	loc := bundle.NewLocalizer("en")
	assert.Equal(t, "Hello", loc.Get("hello"))
}
