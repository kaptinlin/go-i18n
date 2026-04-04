package i18n

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectorDetectLocale(t *testing.T) {
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
			name:     "custom query parameter name is used",
			detector: NewDetector(bundle, WithDetectorQueryParam("locale")),
			request:  httptestNewRequest("GET", "/?locale=ja"),
			want:     "ja-JP",
		},
		{
			name:     "falls back to default locale when no source matches",
			detector: NewDetector(bundle),
			request:  httptestNewRequest("GET", "/"),
			want:     "en",
		},
		{
			name:     "query overrides accept language",
			detector: NewDetector(bundle),
			request: func() *http.Request {
				r := httptestNewRequest("GET", "/?lang=ja")
				r.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
				return r
			}(),
			want: "ja-JP",
		},
		{
			name:     "cookie used when query missing",
			detector: NewDetector(bundle),
			request: func() *http.Request {
				r := httptestNewRequest("GET", "/")
				r.AddCookie(&http.Cookie{Name: "lang", Value: "zh-CN"})
				return r
			}(),
			want: "zh-Hans",
		},
		{
			name:     "custom header used when enabled",
			detector: NewDetector(bundle),
			request: func() *http.Request {
				r := httptestNewRequest("GET", "/")
				r.Header.Set("X-Language", "ja-JP")
				return r
			}(),
			want: "ja-JP",
		},
		{
			name:     "accept language fallback",
			detector: NewDetector(bundle),
			request: func() *http.Request {
				r := httptestNewRequest("GET", "/")
				r.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
				return r
			}(),
			want: "zh-Hans",
		},
		{
			name:     "invalid explicit locale falls through to accept language",
			detector: NewDetector(bundle),
			request: func() *http.Request {
				r := httptestNewRequest("GET", "/?lang=bad-locale")
				r.Header.Set("Accept-Language", "ja;q=0.9")
				return r
			}(),
			want: "ja-JP",
		},
		{
			name:     "custom priority uses header before query",
			detector: NewDetector(bundle, WithDetectorPriority(DetectorSourceHeader, DetectorSourceQuery, DetectorSourceAccept)),
			request: func() *http.Request {
				r := httptestNewRequest("GET", "/?lang=zh-CN")
				r.Header.Set("X-Language", "ja")
				return r
			}(),
			want: "ja-JP",
		},
		{
			name:     "disabled query source skips query parameter",
			detector: NewDetector(bundle, WithDetectorPriority(DetectorSourceAccept), WithDetectorQueryParam("")),
			request: func() *http.Request {
				r := httptestNewRequest("GET", "/?lang=ja")
				r.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
				return r
			}(),
			want: "zh-Hans",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.detector.DetectLocale(tt.request))
		})
	}
}

func TestLocalizerFromContext(t *testing.T) {
	loc := newTestLocalizer()

	ctx := ContextWithLocalizer(context.Background(), loc)
	got, ok := LocalizerFromContext(ctx)
	assert.True(t, ok)
	assert.Same(t, loc, got)
}

func TestLocalizerFromContextMissing(t *testing.T) {
	got, ok := LocalizerFromContext(context.Background())
	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestLocalizerFromContextNilStoredValue(t *testing.T) {
	ctx := ContextWithLocalizer(context.Background(), nil)
	got, ok := LocalizerFromContext(ctx)
	assert.False(t, ok)
	assert.Nil(t, got)
}

func httptestNewRequest(method, target string) *http.Request {
	req, err := http.NewRequest(method, target, nil)
	if err != nil {
		panic(err)
	}
	return req
}
