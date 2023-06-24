package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testTranslations = map[string]map[string]string{
	"en": map[string]string{
		"{count, plural, =0 {None} one {1 Apple} other {# Apples}}": "{count, plural, =0 {None} one {1 Apple} other {# Apples}}",
	},

	"zh-Hans": map[string]string{
		// Token-based Translations
		"test_message":  "这是一则测试讯息。",
		"test_template": "你好，{Name}！",
		"test_plural":   "{count, plural, =0 {没有} =1 {只有 1 个} other {有 # 个}}",

		// Text-based Translations.
		"Hello, world!":        "你好，世界！",
		"How are you, {Name}?": "过得如何，{Name}？",
		"Post <verb>":          "发表贴文",
		"Post <noun>":          "文章",

		"{count, plural, =0 {None} one {1 Apple} other {# Apples}}":         "{count, plural, =0 {没有苹果} =1 {1 颗苹果} other {有 # 颗苹果}}",
		"{count, plural, =0 {No Post} one {1 Post} other {# Posts}} <noun>": "{count, plural, =0 {没有文章} =1 {1 篇文章} other {有 # 篇文章}}",
		"{count, plural, =0 {No Post} one {1 Post} other {# Posts}} <verb>": "{count, plural, =0 {没有发表} =1 {1 篇发表} other {有 # 篇发表}}",

		"Post": "THIS_SHOULD_NOT_BE_USED",
		"{count, plural, =0 {No Post} one {1 Post} other {# Posts}}": "THIS_SHOULD_NOT_BE_USED",
	},

	"ja-JP": map[string]string{
		// Token-based Translations
		"test_message":  "これはテストメッセージです。",
		"test_template": "こんにちは、{Name}！",
		"test_plural":   "{count, plural, =0 {なし} one {1 つだけ} other {# 个あります}}",
	},

	"ko-KR": map[string]string{
		// Token-based Translations
		"test_message":  "이것은 테스트 메시지입니다.",
		"test_template": "안녕하세요, {Name} 님!",
		"test_plural":   "{count, plural, =0 {없음} one {1 개} other {# 개가 있음}}",

		// Text-based Translations.
		"Hello, world!":        "안녕하세요, 세상!",
		"How are you, {Name}?": "{Name} 님, 어떻게 지내세요?",
		"Post <verb>":          "메시지 게시",
		"Post <noun>":          "기사",
	},
}

func newTestLocalizer() *Localizer {
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans", "ja-JP", "ko-KR"),
	)
	bundle.LoadMessages(testTranslations)
	return bundle.NewLocalizer("zh-Hans")
}

func TestLocalizer(t *testing.T) {
	assert := assert.New(t)
	localizer := newTestLocalizer()

	assert.Equal("zh-Hans", localizer.Locale())
}

func TestTokenString(t *testing.T) {
	assert := assert.New(t)
	localizer := newTestLocalizer()

	assert.Equal("这是一则测试讯息。", localizer.Get("test_message"))
	assert.Equal("not_exists_message", localizer.Get("not_exists_message"))
}

func TestTokenVars(t *testing.T) {
	assert := assert.New(t)
	localizer := newTestLocalizer()

	assert.Equal("你好，Yami！", localizer.Get("test_template", Vars{
		"Name": "Yami",
	}))
}

func TestTokenPlural(t *testing.T) {
	assert := assert.New(t)
	localizer := newTestLocalizer()

	assert.Equal("没有", localizer.Get("test_plural", Vars{
		"count": 0,
	}))
	assert.Equal("只有 1 个", localizer.Get("test_plural", Vars{
		"count": 1,
	}))
	assert.Equal("有 2 个", localizer.Get("test_plural", Vars{
		"count": 2,
	}))
}

func TestTextString(t *testing.T) {
	assert := assert.New(t)
	localizer := newTestLocalizer()

	assert.Equal("你好，世界！", localizer.Get("Hello, world!"))
}

func TestTextStringRaw(t *testing.T) {
	assert := assert.New(t)
	localizer := newTestLocalizer()

	assert.Equal("I'm fine thank you!", localizer.Get("I'm fine thank you!"))
}

func TestTextVars(t *testing.T) {
	assert := assert.New(t)
	localizer := newTestLocalizer()

	assert.Equal("过得如何，Yami？", localizer.Get("How are you, {Name}?", Vars{
		"Name": "Yami",
	}))
}

func TestTextVarsRaw(t *testing.T) {
	assert := assert.New(t)
	localizer := newTestLocalizer()

	assert.Equal("I'm fine, thanks to Yami!", localizer.Get("I'm fine, thanks to {Name}!", Vars{
		"Name": "Yami",
	}))
}

func TestTextPlural(t *testing.T) {
	assert := assert.New(t)
	localizer := newTestLocalizer()

	assert.Equal("没有苹果", localizer.Get("{count, plural, =0 {None} one {1 Apple} other {# Apples}}", Vars{
		"count": 0,
	}))
	assert.Equal("1 颗苹果", localizer.Get("{count, plural, =0 {None} one {1 Apple} other {# Apples}}", Vars{
		"count": 1,
	}))
	assert.Equal("有 2 颗苹果", localizer.Get("{count, plural, =0 {None} one {1 Apple} other {# Apples}}", Vars{
		"count": 2,
	}))
}

func TestTextStringContext(t *testing.T) {
	assert := assert.New(t)
	localizer := newTestLocalizer()

	assert.Equal("发表贴文", localizer.GetX("Post", "verb"))
	assert.Equal("文章", localizer.GetX("Post", "noun"))
}

func TestTextPluralContext(t *testing.T) {
	assert := assert.New(t)
	localizer := newTestLocalizer()

	assert.Equal("没有文章", localizer.GetX("{count, plural, =0 {No Post} one {1 Post} other {# Posts}}", "noun", Vars{
		"count": 0,
	}))
	assert.Equal("1 篇文章", localizer.GetX("{count, plural, =0 {No Post} one {1 Post} other {# Posts}}", "noun", Vars{
		"count": 1,
	}))
	assert.Equal("有 2 篇文章", localizer.GetX("{count, plural, =0 {No Post} one {1 Post} other {# Posts}}", "noun", Vars{
		"count": 2,
	}))

	assert.Equal("没有发表", localizer.GetX("{count, plural, =0 {No Post} one {1 Post} other {# Posts}}", "verb", Vars{
		"count": 0,
	}))
	assert.Equal("1 篇发表", localizer.GetX("{count, plural, =0 {No Post} one {1 Post} other {# Posts}}", "verb", Vars{
		"count": 1,
	}))
	assert.Equal("有 2 篇发表", localizer.GetX("{count, plural, =0 {No Post} one {1 Post} other {# Posts}}", "verb", Vars{
		"count": 2,
	}))
}

func TestTextFallback(t *testing.T) {
	assert := assert.New(t)
	bundle := NewBundle(
		WithDefaultLocale("zh-Hans"),
		WithLocales("en", "zh-Hans", "ja-JP", "ko-KR"),
		WithFallback(map[string][]string{
			"ja-JP": []string{"ko-KR"},
		}),
	)
	bundle.LoadMessages(testTranslations)
	localizer := bundle.NewLocalizer("ja-JP")

	// Test ja-JP
	assert.Equal("これはテストメッセージです。", localizer.Get("test_message"))
	assert.Equal("こんにちは、Yami！", localizer.Get("test_template", Vars{
		"Name": "Yami",
	}))
	assert.Equal("なし", localizer.Get("test_plural", Vars{
		"count": 0,
	}))

	// Test ja-JP -> ko-KR fallback
	assert.Equal("안녕하세요, 세상!", localizer.Get("Hello, world!"))
	assert.Equal("Yami 님, 어떻게 지내세요?", localizer.Get("How are you, {Name}?", Vars{
		"Name": "Yami",
	}))
	assert.Equal("메시지 게시", localizer.GetX("Post", "verb"))

	// Test ja-JP -> zh-CN fallback
	assert.Equal("没有苹果", localizer.Get("{count, plural, =0 {None} one {1 Apple} other {# Apples}}", Vars{
		"count": 0,
	}))
	assert.Equal("1 颗苹果", localizer.Get("{count, plural, =0 {None} one {1 Apple} other {# Apples}}", Vars{
		"count": 1,
	}))
	assert.Equal("有 2 颗苹果", localizer.Get("{count, plural, =0 {None} one {1 Apple} other {# Apples}}", Vars{
		"count": 2,
	}))

	// Test nil fallback
	assert.Equal("Ni hao", localizer.Get("Ni hao"))
}

func TestTextFallbackResursive(t *testing.T) {
	assert := assert.New(t)
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans", "ja-JP", "ko-KR"),
		WithFallback(map[string][]string{
			"ja-JP": []string{"ko-KR"},
			"ko-KR": []string{"zh-Hans"},
		}))
	bundle.LoadMessages(testTranslations)
	localizer := bundle.NewLocalizer("ja-JP")

	// Test ja-JP -> ko-KR -> zh-CN fallback
	assert.Equal("1 颗苹果", localizer.Get("{count, plural, =0 {None} one {1 Apple} other {# Apples}}", Vars{
		"count": 1,
	}))
}
