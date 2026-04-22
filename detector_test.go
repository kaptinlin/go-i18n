package i18n

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectorDetectLocale(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans", "ja-JP"),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello"},
		"zh-Hans": {"hello": "你好"},
		"ja-JP":   {"hello": "こんにちは"},
	}))

	tests := []struct {
		name     string
		detector *Detector
		request  *http.Request
		want     string
	}{
		{
			name:     "custom cookie name is used",
			detector: NewDetector(bundle, WithDetectorCookieName("locale")),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/")
				r.AddCookie(&http.Cookie{Name: "locale", Value: "ja"})
				return r
			}(),
			want: "ja-JP",
		},
		{
			name:     "disabled cookie source skips cookie value",
			detector: NewDetector(bundle, WithDetectorPriority(DetectorSourceCookie, DetectorSourceAccept), WithDetectorCookieName("")),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/")
				r.AddCookie(&http.Cookie{Name: "lang", Value: "ja"})
				r.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
				return r
			}(),
			want: "zh-Hans",
		},
		{
			name:     "disabled header source skips explicit header",
			detector: NewDetector(bundle, WithDetectorPriority(DetectorSourceHeader, DetectorSourceAccept), WithDetectorHeaderName("")),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/")
				r.Header.Set("X-Language", "ja")
				r.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
				return r
			}(),
			want: "zh-Hans",
		},
		{
			name:     "custom query parameter name is used",
			detector: NewDetector(bundle, WithDetectorQueryParam("locale")),
			request:  httptestNewRequest(t, "GET", "/?locale=ja"),
			want:     "ja-JP",
		},
		{
			name:     "falls back to default locale when no source matches",
			detector: NewDetector(bundle),
			request:  httptestNewRequest(t, "GET", "/"),
			want:     "en",
		},
		{
			name:     "query overrides accept language",
			detector: NewDetector(bundle),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/?lang=ja")
				r.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
				return r
			}(),
			want: "ja-JP",
		},
		{
			name:     "cookie used when query missing",
			detector: NewDetector(bundle),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/")
				r.AddCookie(&http.Cookie{Name: "lang", Value: "zh-CN"})
				return r
			}(),
			want: "zh-Hans",
		},
		{
			name:     "custom header used when enabled",
			detector: NewDetector(bundle),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/")
				r.Header.Set("X-Language", "ja-JP")
				return r
			}(),
			want: "ja-JP",
		},
		{
			name:     "accept language fallback",
			detector: NewDetector(bundle),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/")
				r.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
				return r
			}(),
			want: "zh-Hans",
		},
		{
			name:     "invalid explicit locale falls through to accept language",
			detector: NewDetector(bundle),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/?lang=bad-locale")
				r.Header.Set("Accept-Language", "ja;q=0.9")
				return r
			}(),
			want: "ja-JP",
		},
		{
			name:     "custom priority uses header before query",
			detector: NewDetector(bundle, WithDetectorPriority(DetectorSourceHeader, DetectorSourceQuery, DetectorSourceAccept)),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/?lang=zh-CN")
				r.Header.Set("X-Language", "ja")
				return r
			}(),
			want: "ja-JP",
		},
		{
			name:     "disabled query source skips query parameter",
			detector: NewDetector(bundle, WithDetectorPriority(DetectorSourceAccept), WithDetectorQueryParam("")),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/?lang=ja")
				r.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
				return r
			}(),
			want: "zh-Hans",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.detector.DetectLocale(tt.request))
		})
	}
}

func TestDetectorExplicitLocaleFallsBackWhenMatchedLocaleHasNoLoadedTranslations(t *testing.T) {
	t.Parallel()
	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en": {"hello": "Hello"},
	}))

	detector := NewDetector(bundle)
	request := httptestNewRequest(t, "GET", "/?lang=zh-CN")

	assert.Equal(t, "en", detector.DetectLocale(request))
}

func TestDetectorPriorityIgnoresInvalidSources(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello"},
		"zh-Hans": {"hello": "你好"},
	}))

	detector := NewDetector(
		bundle,
		WithDetectorPriority(DetectorSource("invalid"), DetectorSourceHeader),
		WithDetectorHeaderName("X-Locale"),
	)

	request := httptestNewRequest(t, "GET", "/?lang=en")
	request.Header.Set("X-Locale", "zh-CN")

	assert.Equal(t, "zh-Hans", detector.DetectLocale(request))
}

func TestLocalizerFromContext(t *testing.T) {
	t.Parallel()

	loc := newTestLocalizer(t)

	ctx := ContextWithLocalizer(context.Background(), loc)
	got, ok := LocalizerFromContext(ctx)
	assert.True(t, ok)
	assert.Same(t, loc, got)
}

func TestLocalizerFromContextMissing(t *testing.T) {
	t.Parallel()

	got, ok := LocalizerFromContext(context.Background())
	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestLocalizerFromContextNilStoredValue(t *testing.T) {
	t.Parallel()

	ctx := ContextWithLocalizer(context.Background(), nil)
	got, ok := LocalizerFromContext(ctx)
	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestLocalizerFromContextNilContext(t *testing.T) {
	t.Parallel()

	var ctx context.Context
	got, ok := LocalizerFromContext(ctx)
	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestDetectorPriorityWithOnlyInvalidSourcesKeepsDefaults(t *testing.T) {
	t.Parallel()

	bundle := NewBundle(
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello"},
		"zh-Hans": {"hello": "你好"},
	}))

	detector := NewDetector(bundle, WithDetectorPriority(DetectorSource("bad")))
	request := httptestNewRequest(t, "GET", "/")
	request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

	assert.Equal(t, "zh-Hans", detector.DetectLocale(request))
}

func httptestNewRequest(tb testing.TB, method, target string) *http.Request {
	tb.Helper()

	req, err := http.NewRequest(method, target, nil)
	require.NoError(tb, err)
	return req
}
