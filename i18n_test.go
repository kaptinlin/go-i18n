package i18n

import (
	"embed"
	"testing"

	mf "github.com/kaptinlin/messageformat-go/v1"
	toml "github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"

	yaml "github.com/goccy/go-yaml"
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

func TestSupportedLocales(t *testing.T) {
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans", "ja-JP"),
	)
	langs := bundle.SupportedLocales()
	assert.Len(t, langs, 3)
	assert.Equal(t, "en", langs[0].String())
}

func TestSupportedLocalesReturnsCopy(t *testing.T) {
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans", "ja-JP"),
	)

	langs := bundle.SupportedLocales()
	langs[0] = language.Japanese

	assert.Equal(t, "en", bundle.SupportedLocales()[0].String())
}

func TestWithFallbackClonesInput(t *testing.T) {
	fallbacks := map[string][]string{
		"ja-JP": {"ko-KR"},
	}

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "ja-JP", "ko-KR", "zh-Hans"),
		WithFallback(fallbacks),
	)

	fallbacks["ja-JP"][0] = "zh-Hans"
	fallbacks["ja-JP"] = append(fallbacks["ja-JP"], "en")

	assert.Equal(t, []string{"ko-KR"}, bundle.fallbacks["ja-JP"])
}

func TestWithMessageFormatOptionsClonesInput(t *testing.T) {
	options := &mf.MessageFormatOptions{
		Strict: true,
		CustomFormatters: map[string]any{
			"upper": func(value any, locale string, arg *string) any {
				return value
			},
		},
	}

	bundle := NewBundle(WithDefaultLocale("en"), WithMessageFormatOptions(options))

	options.Strict = false
	options.CustomFormatters["lower"] = func(value any, locale string, arg *string) any {
		return value
	}

	if assert.NotNil(t, bundle.mfOptions) {
		assert.True(t, bundle.mfOptions.Strict)
		assert.Len(t, bundle.mfOptions.CustomFormatters, 1)
		_, ok := bundle.mfOptions.CustomFormatters["upper"]
		assert.True(t, ok)
	}
}

func TestWithMessageFormatOptionsNil(t *testing.T) {
	bundle := NewBundle(WithDefaultLocale("en"), WithMessageFormatOptions(nil))

	assert.Nil(t, bundle.mfOptions)
}

func TestWithCustomFormattersClonesInput(t *testing.T) {
	formatters := map[string]any{
		"upper": func(value any, locale string, arg *string) any {
			return value
		},
	}

	bundle := NewBundle(WithDefaultLocale("en"), WithCustomFormatters(formatters))

	formatters["lower"] = func(value any, locale string, arg *string) any {
		return value
	}

	if assert.NotNil(t, bundle.mfOptions) {
		assert.Len(t, bundle.mfOptions.CustomFormatters, 1)
		_, ok := bundle.mfOptions.CustomFormatters["upper"]
		assert.True(t, ok)
	}
}

func TestHas(t *testing.T) {
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
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	err := bundle.LoadMessages(map[string]map[string]string{
		"en":      {"b": "B", "a": "A", "shared": "Shared"},
		"zh-Hans": {"b": "乙"},
	})
	assert.NoError(t, err)

	assert.Equal(t, []string{"a", "b", "shared"}, bundle.Keys("en"))
	assert.Equal(t, []string{"b"}, bundle.Keys("zh-CN"))
	assert.Nil(t, bundle.Keys("af"))
}

func TestKeysDoNotFallBackToDefaultForMatchedLocaleWithoutDirectTranslations(t *testing.T) {
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

	assert.Equal(t, []string{"c"}, bundle.Keys("ja-JP"))
}

func TestIsLanguageSupported(t *testing.T) {
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans", "ja-JP"),
	)

	assert.True(t, bundle.IsLanguageSupported(language.English))
	assert.True(t, bundle.IsLanguageSupported(language.Japanese))
	assert.False(t, bundle.IsLanguageSupported(language.Afrikaans))
}

func TestMatchExactLocaleNoMatch(t *testing.T) {
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	// A locale not in the bundle should return empty string.
	got := bundle.matchExactLocale("af")
	assert.Empty(t, got)
}

func TestNewLocalizerMatchedButNoTranslations(t *testing.T) {
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

func TestNewLocalizerNoMatchFallsToDefault(t *testing.T) {
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
	// When no default locale is set, the first locale should be used.
	bundle := NewBundle(WithLocales("fr", "de"))
	assert.Equal(t, "fr", bundle.defaultLocale)
}

func TestWithLocalesInvalidIgnored(t *testing.T) {
	// Invalid locale strings should be silently ignored.
	bundle := NewBundle(WithLocales("en", "???invalid???"))
	assert.Len(t, bundle.languages, 1)
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

func TestParseTranslationSuccess(t *testing.T) {
	assert := assert.New(t)
	bundle := NewBundle(WithDefaultLocale("en"))

	pt, err := bundle.parseTranslation("en", "test_key", "Hello, {name}!")
	assert.NoError(err)
	assert.NotNil(pt)
	assert.Equal("test_key", pt.name)
	assert.Equal("en", pt.locale)
	assert.Equal("Hello, {name}!", pt.text)
	assert.NotNil(pt.format) // MessageFormat should be compiled successfully
}

func TestParseTranslationInvalidMessageFormat(t *testing.T) {
	assert := assert.New(t)
	bundle := NewBundle(WithDefaultLocale("en"))

	// Invalid MessageFormat syntax (unclosed brace)
	pt, err := bundle.parseTranslation("en", "test_key", "Hello, {name")
	assert.ErrorIs(err, ErrMessageFormatCompilation)
	assert.ErrorContains(err, `locale "en"`)
	assert.ErrorContains(err, `key "test_key"`)
	assert.NotNil(pt) // Returns pt with raw text even on error
	assert.Equal("Hello, {name", pt.text)
	assert.Nil(pt.format) // No compiled format available
}

func TestParseTranslationEmptyMessageFormat(t *testing.T) {
	assert := assert.New(t)
	bundle := NewBundle(WithDefaultLocale("en"))

	// Empty message should compile successfully (just no formatting)
	pt, err := bundle.parseTranslation("en", "test_key", "")
	assert.NoError(err)
	assert.NotNil(pt)
	assert.Equal("", pt.text)
}

func TestParseTranslationComplexMessageFormat(t *testing.T) {
	assert := assert.New(t)
	bundle := NewBundle(WithDefaultLocale("en"))

	// Complex MessageFormat with plural
	pt, err := bundle.parseTranslation("en", "test_key", "{count, plural, =0 {none} one {# item} other {# items}}")
	assert.NoError(err)
	assert.NotNil(pt)
	assert.NotNil(pt.format)
}

func TestParseTranslationInvalidLocale(t *testing.T) {
	assert := assert.New(t)
	bundle := NewBundle(WithDefaultLocale("en"))

	pt, err := bundle.parseTranslation("???invalid???", "test_key", "Hello")
	assert.ErrorIs(err, ErrMessageFormatCompilation)
	assert.ErrorContains(err, `locale "???invalid???"`)
	assert.ErrorContains(err, `parse locale "???invalid???"`)
	assert.NotNil(pt)
	assert.Equal("Hello", pt.text)
	assert.Nil(pt.format)
}

func TestLookupInvalidRuntimeLocaleFallsBackToRawText(t *testing.T) {
	assert := assert.New(t)
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	assert.NoError(bundle.LoadMessages(map[string]map[string]string{
		"en": {"valid": "Hello"},
	}))

	loc := &Localizer{bundle: bundle, locale: "???invalid???"}
	r := loc.Lookup("{invalid syntax")
	assert.Equal("{invalid syntax", r.Text)
	assert.Equal("en", r.Locale)
	assert.Equal(TranslationSourceMissing, r.Source)
}

func TestLookupInvalidMessageFormat(t *testing.T) {
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
