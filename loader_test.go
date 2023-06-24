package i18n

import (
	"testing"

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
