package i18n

import (
	"embed"
	"testing"

	"github.com/google/go-cmp/cmp"
	mf "github.com/kaptinlin/messageformat-go/v1"
	toml "github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"

	yaml "github.com/goccy/go-yaml"
)

//go:embed test/*.json
var testTranslationFS embed.FS

func TestLoadMessages(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	bundle := NewBundle()
	assert.NotNil(t, bundle)
	assert.Equal(t, "en", bundle.NewLocalizer().Locale())
	assert.Len(t, bundle.SupportedLocales(), 1)
}

func TestSupportedLocales(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans", "ja-JP"),
	)
	langs := bundle.SupportedLocales()
	assert.Len(t, langs, 3)
	assert.Equal(t, "en", langs[0].String())
}

func TestSupportedLocalesReturnsCopy(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans", "ja-JP"),
	)

	langs := bundle.SupportedLocales()
	langs[0] = language.Japanese

	assert.Equal(t, "en", bundle.SupportedLocales()[0].String())
}

func TestWithFallbackClonesInput(t *testing.T) {
	t.Parallel()

	fallbacks := map[string][]string{
		"ja-JP": {"ko-KR"},
	}

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "ja-JP", "ko-KR", "zh-Hans"),
		WithFallback(fallbacks),
	)
	assert.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"shared": "English"},
		"ko-KR":   {"shared": "Korean"},
		"zh-Hans": {"shared": "Chinese"},
		"ja-JP":   {},
	}))

	fallbacks["ja-JP"][0] = "zh-Hans"
	fallbacks["ja-JP"] = append(fallbacks["ja-JP"], "en")

	assert.Equal(t, "Korean", bundle.NewLocalizer("ja-JP").Get("shared"))
}

func TestWithFallbackNilClearsFallbacks(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "ja-JP", "ko-KR"),
		WithFallback(map[string][]string{"ja-JP": {"ko-KR"}}),
		WithFallback(nil),
	)
	assert.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":    {"shared": "English"},
		"ko-KR": {"shared": "Korean"},
		"ja-JP": {},
	}))

	assert.Equal(t, "English", bundle.NewLocalizer("ja-JP").Get("shared"))
}

func TestWithMessageFormatOptionsClonesInput(t *testing.T) {
	t.Parallel()

	options := &mf.MessageFormatOptions{
		CustomFormatters: map[string]any{
			"upper": func(value any, locale string, arg *string) any {
				return "ORIGINAL"
			},
		},
	}

	bundle := NewBundle(WithDefaultLocale("en"), WithMessageFormatOptions(options))

	options.CustomFormatters["upper"] = func(value any, locale string, arg *string) any {
		return "MUTATED"
	}

	result, err := bundle.NewLocalizer("en").Format("Hello, {name, upper}!", Vars{"name": "ignored"})
	assert.NoError(t, err)
	assert.Equal(t, "Hello, ORIGINAL!", result)
}

func TestWithCustomFormattersClonesInput(t *testing.T) {
	t.Parallel()

	formatters := map[string]any{
		"upper": func(value any, locale string, arg *string) any {
			return "ORIGINAL"
		},
	}

	bundle := NewBundle(WithDefaultLocale("en"), WithCustomFormatters(formatters))

	formatters["upper"] = func(value any, locale string, arg *string) any {
		return "MUTATED"
	}

	result, err := bundle.NewLocalizer("en").Format("Hello, {name, upper}!", Vars{"name": "ignored"})
	assert.NoError(t, err)
	assert.Equal(t, "Hello, ORIGINAL!", result)
}

func TestHas(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	err := bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello", "fallback": "Fallback"},
		"zh-Hans": {"hello": "你好"},
	})
	assert.NoError(t, err)

	assert.True(t, bundle.Has("zh-Hans", "hello"))
	assert.False(t, bundle.Has("zh-CN", "fallback"))
	assert.False(t, bundle.Has("zh-Hans", "missing"))
	assert.False(t, bundle.Has("af", "hello"))
}

func TestHasDoesNotFallBackToDefaultForMatchedLocaleWithoutDirectTranslations(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	err := bundle.LoadMessages(map[string]map[string]string{
		"en": {"fallback": "Fallback"},
	})
	assert.NoError(t, err)

	assert.False(t, bundle.Has("zh-CN", "fallback"))
}

func TestHasIgnoresFallbackKeys(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans", "ja-JP"),
		WithFallback(map[string][]string{
			"ja-JP": {"zh-Hans"},
		}),
	)
	err := bundle.LoadMessages(map[string]map[string]string{
		"en":      {"shared": "English"},
		"zh-Hans": {"shared": "中文"},
		"ja-JP":   {"hello": "こんにちは"},
	})
	assert.NoError(t, err)

	assert.True(t, bundle.Has("ja-JP", "hello"))
	assert.False(t, bundle.Has("ja-JP", "shared"))
}

func TestKeys(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	err := bundle.LoadMessages(map[string]map[string]string{
		"en":      {"b": "B", "a": "A", "shared": "Shared"},
		"zh-Hans": {"b": "乙"},
	})
	assert.NoError(t, err)

	if diff := cmp.Diff([]string{"a", "b", "shared"}, bundle.Keys("en")); diff != "" {
		t.Errorf("keys for en mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"b"}, bundle.Keys("zh-CN")); diff != "" {
		t.Errorf("keys for zh-CN mismatch (-want +got):\n%s", diff)
	}
	assert.Nil(t, bundle.Keys("af"))
}

func TestKeysDoNotFallBackToDefaultForMatchedLocaleWithoutDirectTranslations(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	err := bundle.LoadMessages(map[string]map[string]string{
		"en": {"fallback": "Fallback"},
	})
	assert.NoError(t, err)

	assert.Nil(t, bundle.Keys("zh-CN"))
}

func TestKeysIgnoreFallbackKeys(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans", "ja-JP"),
		WithFallback(map[string][]string{
			"ja-JP": {"zh-Hans"},
		}),
	)
	err := bundle.LoadMessages(map[string]map[string]string{
		"en":      {"a": "A", "b": "B"},
		"zh-Hans": {"b": "乙"},
		"ja-JP":   {"c": "C"},
	})
	assert.NoError(t, err)

	if diff := cmp.Diff([]string{"c"}, bundle.Keys("ja-JP")); diff != "" {
		t.Errorf("keys for ja-JP mismatch (-want +got):\n%s", diff)
	}
}

func TestNewLocalizerKeepsFirstLoadedMatchWhenRequestedLocaleLacksDirectTranslations(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans", "ja-JP"),
	)
	assert.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello"},
		"zh-Hans": {"hello": "你好"},
	}))

	loc := bundle.NewLocalizer("ja-JP", "zh-CN")
	assert.Equal(t, "zh-Hans", loc.Locale())
}

func TestIsLanguageSupported(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans", "ja-JP"),
	)

	assert.True(t, bundle.IsLanguageSupported(language.English))
	assert.True(t, bundle.IsLanguageSupported(language.Japanese))
	assert.False(t, bundle.IsLanguageSupported(language.Afrikaans))
}

func TestNewLocalizerMatchedButNoTranslations(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	// Load translations only for en, not zh-Hans.
	err := bundle.LoadMessages(map[string]map[string]string{
		"en": {"hello": "Hello"},
	})
	assert.NoError(t, err)

	// zh-Hans matches but has no parsedTranslations entry,
	// so NewLocalizer should fall through to default.
	loc := bundle.NewLocalizer("zh-Hans")
	assert.Equal(t, "en", loc.Locale())
}

func TestNewLocalizerMatchedLocaleWithoutLoadedTranslationsFallsToDefault(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	assert.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en": {"hello": "Hello"},
	}))

	loc := bundle.NewLocalizer("zh-CN")
	assert.Equal(t, "en", loc.Locale())
	assert.Equal(t, "Hello", loc.Get("hello"))
}

func TestNewLocalizerNoMatchFallsToDefault(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en"),
	)
	err := bundle.LoadMessages(map[string]map[string]string{
		"en": {"hello": "Hello"},
	})
	assert.NoError(t, err)

	loc := bundle.NewLocalizer("af", "xx")
	assert.Equal(t, "en", loc.Locale())
}

func TestNewLocalizerUsesLocaleMatchingForLoadedTranslations(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	err := bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello"},
		"zh-Hans": {"hello": "你好"},
	})
	assert.NoError(t, err)

	loc := bundle.NewLocalizer("zh-CN")
	assert.Equal(t, "zh-Hans", loc.Locale())
	assert.Equal(t, "你好", loc.Get("hello"))
}

func TestNewBundleDefaultLocaleFromLocales(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(WithLocales("fr", "de"))
	assert.Equal(t, "fr", bundle.NewLocalizer().Locale())
}

func TestWithLocalesInvalidIgnored(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(WithLocales("en", "???invalid???"))
	assert.Len(t, bundle.SupportedLocales(), 1)
}

func TestCircularFallback(t *testing.T) {
	t.Parallel()

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

func TestLookupInvalidMessageFormat(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	err := bundle.LoadMessages(map[string]map[string]string{
		"en": {"valid": "Hello"},
	})
	assert.NoError(err)

	localizer := bundle.NewLocalizer("zh-Hans")

	// Trigger runtime fallback with invalid MessageFormat syntax.
	// The key itself will be used as text; compilation fails gracefully.
	r := localizer.Lookup("{invalid syntax")
	assert.Equal("{invalid syntax", r.Text)
	assert.Equal("en", r.Locale)
	assert.Equal(TranslationSourceMissing, r.Source)
}
