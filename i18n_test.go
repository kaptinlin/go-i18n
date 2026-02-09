package i18n

import (
	"embed"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"

	yaml "gopkg.in/yaml.v3"
)

//go:embed test/*.json
var testTranslationFS embed.FS

func TestLoadMessages(t *testing.T) {
	assert := assert.New(t)

	bundle := NewBundle(
		WithDefaultLocale("zh-Hans"),
		WithLocales("zh-Hans", "ja-JP", "ko-KR"),
	)
	assert.NoError(bundle.LoadMessages(map[string]map[string]string{
		"zh-Hans": {
			"test_message": "这是一则测试讯息。",
		},

		"ja-JP": {
			"test_message": "これはテストメッセージです。",
		},

		"ko-KR": {
			"test_message": "이것은 테스트 메시지입니다.",
		},
	}))
	localizer := bundle.NewLocalizer("zh-Hans")

	assert.Equal("这是一则测试讯息。", localizer.Get("test_message"))
	assert.Equal("not_exists_message", localizer.Get("not_exists_message"))
}

func TestUnmarshaler(t *testing.T) {
	assert := assert.New(t)

	bundle := NewBundle(
		WithDefaultLocale("zh-Hans"),
		WithLocales("zh-Hans"),
		WithUnmarshaler(yaml.Unmarshal),
	)
	assert.NoError(bundle.LoadFiles("test/zh-Hans.yml"))

	localizer := bundle.NewLocalizer("zh-Hans")
	assert.Equal("讯息 A", localizer.Get("message_a"))
}

func TestTomlUnmarshaler(t *testing.T) {
	assert := assert.New(t)

	bundle := NewBundle(
		WithDefaultLocale("zh-Hans"),
		WithLocales("zh-Hans"),
		WithUnmarshaler(toml.Unmarshal),
	)
	assert.NoError(bundle.LoadFiles("test/zh-Hans.toml"))

	localizer := bundle.NewLocalizer("zh-Hans")
	assert.Equal("讯息 A", localizer.Get("message_a"))
}

func TestNewBundleNoOptions(t *testing.T) {
	assert := assert.New(t)

	// NewBundle with no options should not panic and default to English.
	bundle := NewBundle()
	assert.NotNil(bundle)
	assert.Equal("en", bundle.defaultLocale)
	assert.Len(bundle.languages, 1)
}

func TestTrimContext(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Post <verb>", "Post"},
		{"Post <noun>", "Post"},
		{"{count, plural, =0 {No Post}} <noun>", "{count, plural, =0 {No Post}}"},
		{"Hello, world!", "Hello, world!"},
		{"no context", "no context"},
		{"", ""},
		{"<only>", "<only>"},        // no space before <
		{"test < spaced >", "test"}, // space before <
		{"a <b> <c>", "a <b>"},      // only last context removed
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, trimContext(tt.input))
		})
	}
}

func TestNameInsensitive(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"zh_CN.music.json", "zh-cn"},
		{"zh-Hans.messages.toml", "zh-hans"},
		{"en_US.yml", "en-us"},
		{"ja-JP", "ja-jp"},
		{"ko_KR.translations.json", "ko-kr"},
		{"path/to/zh_TW.json", "zh-tw"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, nameInsensitive(tt.input))
		})
	}
}

func TestCircularFallback(t *testing.T) {
	assert := assert.New(t)

	// Circular fallback: ja-JP -> ko-KR -> ja-JP should not infinite loop.
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "ja-JP", "ko-KR"),
		WithFallback(map[string][]string{
			"ja-JP": {"ko-KR"},
			"ko-KR": {"ja-JP"},
		}),
	)
	assert.NoError(bundle.LoadMessages(map[string]map[string]string{
		"en":    {"hello": "Hello"},
		"ja-JP": {},
		"ko-KR": {},
	}))

	localizer := bundle.NewLocalizer("ja-JP")
	// Should fall through to default locale without hanging.
	assert.Equal("Hello", localizer.Get("hello"))
}
