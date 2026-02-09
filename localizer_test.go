package i18n

import (
	"fmt"
	"strings"
	"testing"

	mf "github.com/kaptinlin/messageformat-go/v1"
	"github.com/stretchr/testify/assert"
)

var testTranslations = map[string]map[string]string{
	"en": {
		"{count, plural, =0 {None} one {1 Apple} other {# Apples}}": "{count, plural, =0 {None} one {1 Apple} other {# Apples}}",
	},

	"zh-Hans": {
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

	"ja-JP": {
		// Token-based Translations
		"test_message":  "これはテストメッセージです。",
		"test_template": "こんにちは、{Name}！",
		"test_plural":   "{count, plural, =0 {なし} one {1 つだけ} other {# 个あります}}",
	},

	"ko-KR": {
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
	if err := bundle.LoadMessages(testTranslations); err != nil {
		panic(err)
	}
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
			"ja-JP": {"ko-KR"},
		}),
	)
	assert.NoError(bundle.LoadMessages(testTranslations))
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
			"ja-JP": {"ko-KR"},
			"ko-KR": {"zh-Hans"},
		}))
	assert.NoError(bundle.LoadMessages(testTranslations))
	localizer := bundle.NewLocalizer("ja-JP")

	// Test ja-JP -> ko-KR -> zh-CN fallback
	assert.Equal("1 颗苹果", localizer.Get("{count, plural, =0 {None} one {1 Apple} other {# Apples}}", Vars{
		"count": 1,
	}))
}

func TestCustomFormatters(t *testing.T) {
	assert := assert.New(t)

	upperFormatter := func(value interface{}, locale string, arg *string) interface{} {
		return strings.ToUpper(fmt.Sprintf("%v", value))
	}

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithCustomFormatters(map[string]interface{}{
			"upper": upperFormatter,
		}),
	)

	localizer := bundle.NewLocalizer("en")

	result, err := localizer.Format("Hello, {name, upper}!", Vars{
		"name": "world",
	})

	assert.NoError(err)
	assert.Equal("Hello, WORLD!", result)
}

func TestStrictMode(t *testing.T) {
	assert := assert.New(t)

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithStrictMode(true),
	)

	localizer := bundle.NewLocalizer("en")

	result, err := localizer.Format("{count, plural, one {# item} other {# items}}", Vars{
		"count": 1,
	})

	assert.NoError(err)
	assert.Equal("1 item", result)
}

func TestFormatMethod(t *testing.T) {
	assert := assert.New(t)

	bundle := NewBundle(WithDefaultLocale("en"))
	localizer := bundle.NewLocalizer("en")

	result1, err := localizer.Format("Hello, {name}!", Vars{
		"name": "Alice",
	})
	assert.NoError(err)
	assert.Equal("Hello, Alice!", result1)

	result2, err := localizer.Format("{count, plural, =0 {no items} one {# item} other {# items}}", Vars{
		"count": 0,
	})
	assert.NoError(err)
	assert.Equal("no items", result2)

	result3, err := localizer.Format("{count, plural, =0 {no items} one {# item} other {# items}}", Vars{
		"count": 1,
	})
	assert.NoError(err)
	assert.Equal("1 item", result3)

	result4, err := localizer.Format("{count, plural, =0 {no items} one {# item} other {# items}}", Vars{
		"count": 5,
	})
	assert.NoError(err)
	assert.Equal("5 items", result4)
}

func TestMessageFormatOptions(t *testing.T) {
	assert := assert.New(t)

	options := &mf.MessageFormatOptions{
		Strict:   true,
		Currency: "USD",
	}

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithMessageFormatOptions(options),
	)

	localizer := bundle.NewLocalizer("en")

	result, err := localizer.Format("Hello, {name}!", Vars{
		"name": "World",
	})

	assert.NoError(err)
	assert.Equal("Hello, World!", result)
}

func TestGetf(t *testing.T) {
	assert := assert.New(t)
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en"),
	)
	err := bundle.LoadMessages(map[string]map[string]string{
		"en": {"greeting": "Hello, %s! You have %d items."},
	})
	assert.NoError(err)

	loc := bundle.NewLocalizer("en")
	assert.Equal("Hello, Alice! You have 3 items.",
		loc.Getf("greeting", "Alice", 3))
}

func TestGetfMissingKey(t *testing.T) {
	assert := assert.New(t)
	bundle := NewBundle(WithDefaultLocale("en"), WithLocales("en"))
	err := bundle.LoadMessages(map[string]map[string]string{
		"en": {"greeting": "Hello, %s!"},
	})
	assert.NoError(err)

	loc := bundle.NewLocalizer("en")
	// Getf with an existing key applies Sprintf.
	assert.Equal("Hello, Alice!", loc.Getf("greeting", "Alice"))
	// Getf with a missing key: lookup falls through to runtime parse,
	// then Sprintf is applied to the raw name text.
	assert.Equal("no_such_key", loc.Getf("no_such_key"))
}

func TestLocalizeWithoutVars(t *testing.T) {
	assert := assert.New(t)
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en"),
	)
	err := bundle.LoadMessages(map[string]map[string]string{
		"en": {"hello": "Hello, {name}!"},
	})
	assert.NoError(err)

	loc := bundle.NewLocalizer("en")
	// Without vars, raw text is returned even if it has placeholders.
	assert.Equal("Hello, {name}!", loc.Get("hello"))
}

func TestFormatNoVars(t *testing.T) {
	assert := assert.New(t)
	bundle := NewBundle(WithDefaultLocale("en"))
	loc := bundle.NewLocalizer("en")

	result, err := loc.Format("Hello, world!")
	assert.NoError(err)
	assert.Equal("Hello, world!", result)
}

func TestFormatCompileError(t *testing.T) {
	assert := assert.New(t)
	bundle := NewBundle(WithDefaultLocale("en"))
	loc := bundle.NewLocalizer("en")

	// Malformed MessageFormat should return an error.
	_, err := loc.Format("{count, plural, }")
	assert.Error(err)
}

func TestGetRuntimeParsedTranslationCache(t *testing.T) {
	assert := assert.New(t)
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	err := bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello"},
		"zh-Hans": {"hello": "你好"},
	})
	assert.NoError(err)

	loc := bundle.NewLocalizer("zh-Hans")
	// First call: key not in zh-Hans, triggers runtime parse from default.
	assert.Equal("Goodbye", loc.Get("Goodbye"))
	// Second call: should hit runtimeParsedTranslations cache.
	assert.Equal("Goodbye", loc.Get("Goodbye"))
}
