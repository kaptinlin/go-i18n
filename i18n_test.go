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
