package i18n

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	mf "github.com/kaptinlin/messageformat-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testTranslations = map[string]map[string]string{
	"en": {
		"{count, plural, =0 {None} one {1 Apple} other {# Apples}}": "{count, plural, =0 {None} one {1 Apple} other {# Apples}}",
	},

	"zh-Hans": {
		"test_message":  "这是一则测试讯息。",
		"test_template": "你好，{Name}！",
		"test_plural":   "{count, plural, =0 {没有} =1 {只有 1 个} other {有 # 个}}",

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
		"test_message":  "これはテストメッセージです。",
		"test_template": "こんにちは、{Name}！",
		"test_plural":   "{count, plural, =0 {なし} one {1 つだけ} other {# 个あります}}",
	},

	"ko-KR": {
		"test_message":  "이것은 테스트 메시지입니다.",
		"test_template": "안녕하세요, {Name} 님!",
		"test_plural":   "{count, plural, =0 {없음} one {1 개} other {# 개가 있음}}",

		"Hello, world!":        "안녕하세요, 세상!",
		"How are you, {Name}?": "{Name} 님, 어떻게 지내세요?",
		"Post <verb>":          "메시지 게시",
		"Post <noun>":          "기사",
	},
}

func newTestLocalizer(tb testing.TB) *Localizer {
	tb.Helper()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans", "ja-JP", "ko-KR"),
	)
	require.NoError(tb, bundle.LoadMessages(testTranslations))
	return bundle.NewLocalizer("zh-Hans")
}

func TestTokenString(t *testing.T) {
	t.Parallel()

	localizer := newTestLocalizer(t)

	assert.Equal(t, "这是一则测试讯息。", localizer.Get("test_message"))
	assert.Equal(t, "not_exists_message", localizer.Get("not_exists_message"))
}

func TestTokenVars(t *testing.T) {
	t.Parallel()

	localizer := newTestLocalizer(t)

	assert.Equal(t, "你好，Yami！", localizer.Get("test_template", Vars{
		"Name": "Yami",
	}))
}

func TestTokenPlural(t *testing.T) {
	t.Parallel()

	localizer := newTestLocalizer(t)

	assert.Equal(t, "没有", localizer.Get("test_plural", Vars{"count": 0}))
	assert.Equal(t, "只有 1 个", localizer.Get("test_plural", Vars{"count": 1}))
	assert.Equal(t, "有 2 个", localizer.Get("test_plural", Vars{"count": 2}))
}

func TestTextString(t *testing.T) {
	t.Parallel()

	localizer := newTestLocalizer(t)
	assert.Equal(t, "你好，世界！", localizer.Get("Hello, world!"))
}

func TestTextStringRaw(t *testing.T) {
	t.Parallel()

	localizer := newTestLocalizer(t)
	assert.Equal(t, "I'm fine thank you!", localizer.Get("I'm fine thank you!"))
}

func TestTextVars(t *testing.T) {
	t.Parallel()

	localizer := newTestLocalizer(t)
	assert.Equal(t, "过得如何，Yami？", localizer.Get("How are you, {Name}?", Vars{
		"Name": "Yami",
	}))
}

func TestTextVarsRaw(t *testing.T) {
	t.Parallel()

	localizer := newTestLocalizer(t)
	assert.Equal(t, "I'm fine, thanks to Yami!", localizer.Get("I'm fine, thanks to {Name}!", Vars{
		"Name": "Yami",
	}))
}

func TestTextPlural(t *testing.T) {
	t.Parallel()

	localizer := newTestLocalizer(t)

	assert.Equal(t, "没有苹果", localizer.Get("{count, plural, =0 {None} one {1 Apple} other {# Apples}}", Vars{"count": 0}))
	assert.Equal(t, "1 颗苹果", localizer.Get("{count, plural, =0 {None} one {1 Apple} other {# Apples}}", Vars{"count": 1}))
	assert.Equal(t, "有 2 颗苹果", localizer.Get("{count, plural, =0 {None} one {1 Apple} other {# Apples}}", Vars{"count": 2}))
}

func TestTextStringContext(t *testing.T) {
	t.Parallel()

	localizer := newTestLocalizer(t)
	assert.Equal(t, "发表贴文", localizer.GetX("Post", "verb"))
	assert.Equal(t, "文章", localizer.GetX("Post", "noun"))
}

func TestTextPluralContext(t *testing.T) {
	t.Parallel()

	localizer := newTestLocalizer(t)

	assert.Equal(t, "没有文章", localizer.GetX("{count, plural, =0 {No Post} one {1 Post} other {# Posts}}", "noun", Vars{"count": 0}))
	assert.Equal(t, "1 篇文章", localizer.GetX("{count, plural, =0 {No Post} one {1 Post} other {# Posts}}", "noun", Vars{"count": 1}))
	assert.Equal(t, "有 2 篇文章", localizer.GetX("{count, plural, =0 {No Post} one {1 Post} other {# Posts}}", "noun", Vars{"count": 2}))

	assert.Equal(t, "没有发表", localizer.GetX("{count, plural, =0 {No Post} one {1 Post} other {# Posts}}", "verb", Vars{"count": 0}))
	assert.Equal(t, "1 篇发表", localizer.GetX("{count, plural, =0 {No Post} one {1 Post} other {# Posts}}", "verb", Vars{"count": 1}))
	assert.Equal(t, "有 2 篇发表", localizer.GetX("{count, plural, =0 {No Post} one {1 Post} other {# Posts}}", "verb", Vars{"count": 2}))
}

func TestTextFallback(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("zh-Hans"),
		WithLocales("en", "zh-Hans", "ja-JP", "ko-KR"),
		WithFallback(map[string][]string{
			"ja-JP": {"ko-KR"},
		}),
	)
	require.NoError(t, bundle.LoadMessages(testTranslations))
	localizer := bundle.NewLocalizer("ja-JP")

	assert.Equal(t, "これはテストメッセージです。", localizer.Get("test_message"))
	assert.Equal(t, "こんにちは、Yami！", localizer.Get("test_template", Vars{"Name": "Yami"}))
	assert.Equal(t, "なし", localizer.Get("test_plural", Vars{"count": 0}))

	assert.Equal(t, "안녕하세요, 세상!", localizer.Get("Hello, world!"))
	assert.Equal(t, "Yami 님, 어떻게 지내세요?", localizer.Get("How are you, {Name}?", Vars{"Name": "Yami"}))
	assert.Equal(t, "메시지 게시", localizer.GetX("Post", "verb"))

	assert.Equal(t, "没有苹果", localizer.Get("{count, plural, =0 {None} one {1 Apple} other {# Apples}}", Vars{"count": 0}))
	assert.Equal(t, "1 颗苹果", localizer.Get("{count, plural, =0 {None} one {1 Apple} other {# Apples}}", Vars{"count": 1}))
	assert.Equal(t, "有 2 颗苹果", localizer.Get("{count, plural, =0 {None} one {1 Apple} other {# Apples}}", Vars{"count": 2}))

	assert.Equal(t, "Ni hao", localizer.Get("Ni hao"))
}

func TestTextFallbackResursive(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans", "ja-JP", "ko-KR"),
		WithFallback(map[string][]string{
			"ja-JP": {"ko-KR"},
			"ko-KR": {"zh-Hans"},
		}),
	)
	require.NoError(t, bundle.LoadMessages(testTranslations))
	localizer := bundle.NewLocalizer("ja-JP")

	assert.Equal(t, "1 颗苹果", localizer.Get("{count, plural, =0 {None} one {1 Apple} other {# Apples}}", Vars{"count": 1}))
}

func TestCustomFormatters(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithCustomFormatters(map[string]any{
			"upper": func(value any, locale string, arg *string) any {
				return strings.ToUpper(fmt.Sprintf("%v", value))
			},
		}),
	)

	localizer := bundle.NewLocalizer("en")
	result, err := localizer.Format("Hello, {name, upper}!", Vars{
		"name": "world",
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello, WORLD!", result)
}

func TestStrictMode(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithStrictMode(true),
	)

	localizer := bundle.NewLocalizer("en")
	result, err := localizer.Format("{count, plural, one {# item} other {# items}}", Vars{"count": 1})
	require.NoError(t, err)
	assert.Equal(t, "1 item", result)
}

func TestFormatMethod(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(WithDefaultLocale("en"))
	localizer := bundle.NewLocalizer("en")

	result1, err := localizer.Format("Hello, {name}!", Vars{"name": "Alice"})
	require.NoError(t, err)
	assert.Equal(t, "Hello, Alice!", result1)

	result2, err := localizer.Format("{count, plural, =0 {no items} one {# item} other {# items}}", Vars{"count": 0})
	require.NoError(t, err)
	assert.Equal(t, "no items", result2)

	result3, err := localizer.Format("{count, plural, =0 {no items} one {# item} other {# items}}", Vars{"count": 1})
	require.NoError(t, err)
	assert.Equal(t, "1 item", result3)

	result4, err := localizer.Format("{count, plural, =0 {no items} one {# item} other {# items}}", Vars{"count": 5})
	require.NoError(t, err)
	assert.Equal(t, "5 items", result4)
}

func TestMessageFormatOptions(t *testing.T) {
	t.Parallel()

	options := &mf.MessageFormatOptions{
		Strict:   true,
		Currency: "USD",
	}

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithMessageFormatOptions(options),
	)

	localizer := bundle.NewLocalizer("en")
	result, err := localizer.Format("Hello, {name}!", Vars{"name": "World"})
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", result)
}

func TestFormatRespectsStrictMessageFormatOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options *mf.MessageFormatOptions
		want    string
		wantErr bool
	}{
		{
			name: "default mode passes unknown formatter through",
			want: "World",
		},
		{
			name:    "strict mode rejects unknown formatter",
			options: &mf.MessageFormatOptions{Strict: true},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			options := []Option{WithDefaultLocale("en")}
			if tt.options != nil {
				options = append(options, WithMessageFormatOptions(tt.options))
			}

			bundle := NewBundle(options...)
			localizer := bundle.NewLocalizer("en")
			result, err := localizer.Format("{name, upper}", Vars{"name": "World"})
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "compile message")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestLocalizeWithoutVars(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en"),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en": {
			"plain": "Hello, world!",
			"hello": "Hello, {name}!",
		},
	}))

	loc := bundle.NewLocalizer("en")
	assert.Equal(t, "Hello, world!", loc.Get("plain"))
	assert.Equal(t, "Hello, {name}!", loc.Get("hello"))
}

func TestFormatNoVars(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(WithDefaultLocale("en"))
	loc := bundle.NewLocalizer("en")

	result, err := loc.Format("Hello, world!")
	require.NoError(t, err)
	assert.Equal(t, "Hello, world!", result)
}

func TestFormatCompileError(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(WithDefaultLocale("en"))
	loc := bundle.NewLocalizer("en")

	_, err := loc.Format("{count, plural, }")
	require.Error(t, err)
}

func TestFormatInvalidLocalizerLocaleReturnsError(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(WithDefaultLocale("en"))
	loc := &Localizer{bundle: bundle, locale: "???invalid???"}

	_, err := loc.Format("Hello, world!")
	require.Error(t, err)
}

func TestFormatStringerFallbackForNonStringResult(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithCustomFormatters(map[string]any{
			"countWords": func(value any, locale string, arg *string) any {
				return []string{"one", "two"}
			},
		}),
	)
	loc := bundle.NewLocalizer("en")

	result, err := loc.Format("{name, countWords}", Vars{"name": "ignored"})
	require.NoError(t, err)
	assert.Equal(t, "[one two]", result)
}

func TestGetFallsBackToRawTextOnRuntimeFormatError(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(WithDefaultLocale("en"))
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en": {
			"items": "{count, plural, =0 {no items} one {# item} other {# items}}",
		},
	}))

	loc := bundle.NewLocalizer("en")
	assert.Equal(t, "{count, plural, =0 {no items} one {# item} other {# items}}", loc.Get("items", Vars{"count": "oops"}))
}

func TestGetCachesRuntimeFallbackByBehavior(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello"},
		"zh-Hans": {"hello": "你好"},
	}))

	loc := bundle.NewLocalizer("zh-Hans")
	assert.Equal(t, "Goodbye", loc.Get("Goodbye"))
	assert.Equal(t, "Goodbye", loc.Get("Goodbye"))
}

func TestGetRuntimeFallbackIsSafeForConcurrentReads(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello"},
		"zh-Hans": {"hello": "你好"},
	}))

	loc := bundle.NewLocalizer("zh-Hans")

	const goroutines = 32
	const iterations = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			for range iterations {
				assert.Equal(t, "Goodbye", loc.Get("Goodbye"))
			}
		}()
	}
	wg.Wait()
}

func TestLookup(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello", "bye": "Goodbye"},
		"zh-Hans": {"hello": "你好"},
	}))

	loc := bundle.NewLocalizer("zh-Hans")

	r := loc.Lookup("hello")
	assert.Equal(t, "你好", r.Text)
	assert.Equal(t, "zh-Hans", r.Locale)
	assert.Equal(t, TranslationSourceDirect, r.Source)

	r = loc.Lookup("bye")
	assert.Equal(t, "Goodbye", r.Text)
	assert.Equal(t, "en", r.Locale)
	assert.Equal(t, TranslationSourceFallback, r.Source)

	r = loc.Lookup("nonexistent")
	assert.Equal(t, "nonexistent", r.Text)
	assert.Equal(t, "en", r.Locale)
	assert.Equal(t, TranslationSourceMissing, r.Source)
}

func TestLookupContext(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"Post <verb>": "Post", "Post <noun>": "Post"},
		"zh-Hans": {"Post <verb>": "发表", "Post <noun>": "帖子"},
	}))

	loc := bundle.NewLocalizer("zh-Hans")

	r := loc.Lookup("Post <verb>")
	assert.Equal(t, "发表", r.Text)
	assert.Equal(t, "zh-Hans", r.Locale)
	assert.Equal(t, TranslationSourceDirect, r.Source)

	r = loc.Lookup("Post <noun>")
	assert.Equal(t, "帖子", r.Text)
	assert.Equal(t, "zh-Hans", r.Locale)
	assert.Equal(t, TranslationSourceDirect, r.Source)
}

func TestLookupFallbackChain(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans", "ja-JP"),
		WithFallback(map[string][]string{
			"ja-JP": {"zh-Hans"},
		}),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello", "shared_key": "English"},
		"zh-Hans": {"hello": "你好", "shared_key": "Chinese"},
		"ja-JP":   {"hello": "こんにちは"},
	}))

	loc := bundle.NewLocalizer("ja-JP")

	r := loc.Lookup("hello")
	assert.Equal(t, "こんにちは", r.Text)
	assert.Equal(t, "ja-JP", r.Locale)
	assert.Equal(t, TranslationSourceDirect, r.Source)

	r = loc.Lookup("shared_key")
	assert.Equal(t, "Chinese", r.Text)
	assert.Equal(t, "zh-Hans", r.Locale)
	assert.Equal(t, TranslationSourceFallback, r.Source)
}

func TestLookupWithVars(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"greeting": "Hello, {name}!"},
		"zh-Hans": {"greeting": "你好，{name}！"},
	}))

	loc := bundle.NewLocalizer("zh-Hans")

	r := loc.Lookup("greeting", Vars{"name": "World"})
	assert.Equal(t, "你好，World！", r.Text)
	assert.Equal(t, "zh-Hans", r.Locale)
	assert.Equal(t, TranslationSourceDirect, r.Source)

	r = loc.Lookup("unknown", Vars{"name": "Test"})
	assert.Equal(t, "unknown", r.Text)
	assert.Equal(t, "en", r.Locale)
	assert.Equal(t, TranslationSourceMissing, r.Source)
}

func TestLookupDetectFallbackVsDirect(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello", "only_en": "English only"},
		"zh-Hans": {"hello": "你好"},
	}))

	loc := bundle.NewLocalizer("zh-Hans")

	r := loc.Lookup("hello")
	assert.Equal(t, TranslationSourceDirect, r.Source)
	assert.Equal(t, loc.Locale(), r.Locale)

	r = loc.Lookup("only_en")
	assert.Equal(t, TranslationSourceFallback, r.Source)
	assert.NotEqual(t, loc.Locale(), r.Locale)

	r = loc.Lookup("nonexistent")
	assert.Equal(t, TranslationSourceMissing, r.Source)
}
