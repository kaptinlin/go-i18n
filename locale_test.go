package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAcceptLanguage(t *testing.T) {
	assert := assert.New(t)

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("zh-Hans", "ja-JP", "ko-KR"),
	)

	assert.NoError(bundle.LoadMessages(map[string]map[string]string{
		"en": {
			"hello_world": "Hello, world",
		},

		"zh-Hans": {
			"hello_world": "你好，世界",
		},

		"ja-JP": {
			"hello_world": "こんにちは世界",
		},

		"ko-KR": {
			"hello_world": "안녕 세상",
		},
	}))

	localizer := bundle.NewLocalizer(bundle.MatchAvailableLocale("zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,ja;q=0.6"))

	assert.Equal("zh-Hans", localizer.Locale())
	assert.Equal("你好，世界", localizer.Get("hello_world"))

	localizer = bundle.NewLocalizer(bundle.MatchAvailableLocale("en-us;q=0.7,en;q=0.3"))

	assert.Equal("en", localizer.Locale())
	assert.Equal("Hello, world", localizer.Get("hello_world"))

	localizer = bundle.NewLocalizer(bundle.MatchAvailableLocale("ja-JP,ja;q=0.9,en;q=0.8"))

	assert.Equal("ja-JP", localizer.Locale())
	assert.Equal("こんにちは世界", localizer.Get("hello_world"))

	localizer = bundle.NewLocalizer(bundle.MatchAvailableLocale("de;q=0.9,de-DE;q=0.8"))

	assert.Equal("en", localizer.Locale())
	assert.Equal("Hello, world", localizer.Get("hello_world"))
}
