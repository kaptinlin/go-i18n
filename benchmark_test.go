package i18n

import (
	"testing"
)

// BenchmarkLocalizerGet benchmarks simple token-based translation lookup.
func BenchmarkLocalizerGet(b *testing.B) {
	localizer := newTestLocalizer()
	for b.Loop() {
		_ = localizer.Get("test_message")
	}
}

// BenchmarkLocalizerGetMiss benchmarks translation lookup with cache miss (fallback to key).
func BenchmarkLocalizerGetMiss(b *testing.B) {
	localizer := newTestLocalizer()
	for b.Loop() {
		_ = localizer.Get("not_exists_message")
	}
}

// BenchmarkLocalizerGetWithVars benchmarks template-based translation with variable substitution.
func BenchmarkLocalizerGetWithVars(b *testing.B) {
	localizer := newTestLocalizer()
	vars := Vars{"Name": "Yami"}
	for b.Loop() {
		_ = localizer.Get("test_template", vars)
	}
}

// BenchmarkLocalizerGetPlural benchmarks plural form translation.
func BenchmarkLocalizerGetPlural(b *testing.B) {
	localizer := newTestLocalizer()
	vars := Vars{"count": 2}
	for b.Loop() {
		_ = localizer.Get("test_plural", vars)
	}
}

// BenchmarkLocalizerGetPluralZero benchmarks plural form translation with zero value.
func BenchmarkLocalizerGetPluralZero(b *testing.B) {
	localizer := newTestLocalizer()
	vars := Vars{"count": 0}
	for b.Loop() {
		_ = localizer.Get("test_plural", vars)
	}
}

// BenchmarkLocalizerGetX benchmarks context-disambiguated translation.
func BenchmarkLocalizerGetX(b *testing.B) {
	localizer := newTestLocalizer()
	for b.Loop() {
		_ = localizer.GetX("Post", "verb")
	}
}

// BenchmarkLocalizerGetTextBased benchmarks text-based translation lookup.
func BenchmarkLocalizerGetTextBased(b *testing.B) {
	localizer := newTestLocalizer()
	for b.Loop() {
		_ = localizer.Get("Hello, world!")
	}
}

// BenchmarkLocalizerFormat benchmarks direct MessageFormat compilation and formatting.
func BenchmarkLocalizerFormat(b *testing.B) {
	bundle := NewBundle(WithDefaultLocale("en"))
	localizer := bundle.NewLocalizer("en")
	vars := Vars{"name": "Alice"}
	for b.Loop() {
		_, _ = localizer.Format("Hello, {name}!", vars)
	}
}

// BenchmarkLocalizerFormatPlural benchmarks MessageFormat with plural forms.
func BenchmarkLocalizerFormatPlural(b *testing.B) {
	bundle := NewBundle(WithDefaultLocale("en"))
	localizer := bundle.NewLocalizer("en")
	vars := Vars{"count": 5}
	for b.Loop() {
		_, _ = localizer.Format("{count, plural, =0 {no items} one {# item} other {# items}}", vars)
	}
}

// BenchmarkNameInsensitive benchmarks the nameInsensitive function for locale normalization.
func BenchmarkNameInsensitive(b *testing.B) {
	testCases := []string{
		"zh_CN.music.json",
		"zh-Hans.messages.toml",
		"en_US.yml",
		"ja-JP",
		"ko_KR.translations.json",
	}
	for b.Loop() {
		for _, tc := range testCases {
			_ = nameInsensitive(tc)
		}
	}
}

// BenchmarkNameInsensitiveSingle benchmarks nameInsensitive with a single input.
func BenchmarkNameInsensitiveSingle(b *testing.B) {
	for b.Loop() {
		_ = nameInsensitive("zh_CN.music.json")
	}
}

// BenchmarkLoadFiles benchmarks loading translation files.
func BenchmarkLoadFiles(b *testing.B) {
	for b.Loop() {
		bundle := NewBundle(
			WithDefaultLocale("zh-Hans"),
			WithLocales("zh-Hans"),
		)
		_ = bundle.LoadFiles("test/zh-Hans.json", "test/zh_Hans.json", "test/zh-Hans.hello.json")
	}
}

// BenchmarkLoadMessages benchmarks loading translations from Go maps.
func BenchmarkLoadMessages(b *testing.B) {
	translations := map[string]map[string]string{
		"en": {
			"hello":   "Hello",
			"goodbye": "Goodbye",
			"welcome": "Welcome, {name}!",
		},
		"zh-Hans": {
			"hello":   "你好",
			"goodbye": "再见",
			"welcome": "欢迎，{name}！",
		},
	}
	for b.Loop() {
		bundle := NewBundle(
			WithDefaultLocale("en"),
			WithLocales("en", "zh-Hans"),
		)
		_ = bundle.LoadMessages(translations)
	}
}

// BenchmarkNewBundle benchmarks creating a new bundle with options.
func BenchmarkNewBundle(b *testing.B) {
	for b.Loop() {
		_ = NewBundle(
			WithDefaultLocale("en"),
			WithLocales("en", "zh-Hans", "ja-JP", "ko-KR"),
			WithFallback(map[string][]string{
				"ja-JP": {"ko-KR", "zh-Hans"},
			}),
		)
	}
}

// BenchmarkNewLocalizer benchmarks creating a new localizer.
func BenchmarkNewLocalizer(b *testing.B) {
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans", "ja-JP", "ko-KR"),
	)
	_ = bundle.LoadMessages(testTranslations)
	for b.Loop() {
		_ = bundle.NewLocalizer("zh-Hans")
	}
}

// BenchmarkGetf benchmarks sprintf-style formatting.
func BenchmarkGetf(b *testing.B) {
	localizer := newTestLocalizer()
	for b.Loop() {
		_ = localizer.Getf("Hello, %s!", "World")
	}
}

// BenchmarkParallel benchmarks concurrent translation lookups.
func BenchmarkLocalizerGetParallel(b *testing.B) {
	localizer := newTestLocalizer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = localizer.Get("test_message")
		}
	})
}

// BenchmarkLocalizerGetWithVarsParallel benchmarks concurrent template translations.
func BenchmarkLocalizerGetWithVarsParallel(b *testing.B) {
	localizer := newTestLocalizer()
	b.RunParallel(func(pb *testing.PB) {
		vars := Vars{"Name": "Yami"}
		for pb.Next() {
			_ = localizer.Get("test_template", vars)
		}
	})
}
