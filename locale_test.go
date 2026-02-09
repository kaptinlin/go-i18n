package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchAvailableLocale(t *testing.T) {
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("zh-Hans", "ja-JP", "ko-KR"),
	)

	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello_world": "Hello, world"},
		"zh-Hans": {"hello_world": "你好，世界"},
		"ja-JP":   {"hello_world": "こんにちは世界"},
		"ko-KR":   {"hello_world": "안녕 세상"},
	}))

	tests := []struct {
		name       string
		accept     string
		wantLocale string
		wantText   string
	}{
		{
			name:       "Chinese simplified via zh-CN",
			accept:     "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,ja;q=0.6",
			wantLocale: "zh-Hans",
			wantText:   "你好，世界",
		},
		{
			name:       "English via en-us",
			accept:     "en-us;q=0.7,en;q=0.3",
			wantLocale: "en",
			wantText:   "Hello, world",
		},
		{
			name:       "Japanese via ja-JP",
			accept:     "ja-JP,ja;q=0.9,en;q=0.8",
			wantLocale: "ja-JP",
			wantText:   "こんにちは世界",
		},
		{
			name:       "unsupported language falls back to default",
			accept:     "de;q=0.9,de-DE;q=0.8",
			wantLocale: "en",
			wantText:   "Hello, world",
		},
		{
			name:       "invalid header falls back to default",
			accept:     "not-a-valid-header!!!",
			wantLocale: "en",
			wantText:   "Hello, world",
		},
		{
			name:       "empty header falls back to default",
			accept:     "",
			wantLocale: "en",
			wantText:   "Hello, world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locale := bundle.MatchAvailableLocale(tt.accept)
			localizer := bundle.NewLocalizer(locale)

			assert.Equal(t, tt.wantLocale, localizer.Locale())
			assert.Equal(t, tt.wantText, localizer.Get("hello_world"))
		})
	}
}
