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

	bundle := newTestBundle(t,
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
			detector: newTestDetector(t, bundle, WithDetectorCookieName("locale")),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/")
				r.AddCookie(&http.Cookie{Name: "locale", Value: "ja"})
				return r
			}(),
			want: "ja-JP",
		},
		{
			name:     "disabled cookie source skips cookie value",
			detector: newTestDetector(t, bundle, WithDetectorPriority(DetectorSourceCookie, DetectorSourceAccept), WithDetectorCookieName("")),
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
			detector: newTestDetector(t, bundle, WithDetectorPriority(DetectorSourceHeader, DetectorSourceAccept), WithDetectorHeaderName("")),
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
			detector: newTestDetector(t, bundle, WithDetectorQueryParam("locale")),
			request:  httptestNewRequest(t, "GET", "/?locale=ja"),
			want:     "ja-JP",
		},
		{
			name:     "empty query parameter name skips query value",
			detector: newTestDetector(t, bundle, WithDetectorQueryParam("")),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/?lang=ja")
				r.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
				return r
			}(),
			want: "zh-Hans",
		},
		{
			name:     "empty priority keeps default source order",
			detector: newTestDetector(t, bundle, WithDetectorPriority()),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/?lang=ja")
				r.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
				return r
			}(),
			want: "ja-JP",
		},
		{
			name:     "falls back to default locale when no source matches",
			detector: newTestDetector(t, bundle),
			request:  httptestNewRequest(t, "GET", "/"),
			want:     "en",
		},
		{
			name:     "query overrides accept language",
			detector: newTestDetector(t, bundle),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/?lang=ja")
				r.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
				return r
			}(),
			want: "ja-JP",
		},
		{
			name:     "cookie used when query missing",
			detector: newTestDetector(t, bundle),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/")
				r.AddCookie(&http.Cookie{Name: "lang", Value: "zh-CN"})
				return r
			}(),
			want: "zh-Hans",
		},
		{
			name:     "custom header used when enabled",
			detector: newTestDetector(t, bundle),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/")
				r.Header.Set("X-Language", "ja-JP")
				return r
			}(),
			want: "ja-JP",
		},
		{
			name:     "accept language fallback",
			detector: newTestDetector(t, bundle),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/")
				r.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
				return r
			}(),
			want: "zh-Hans",
		},
		{
			name:     "accept language combines all header field values",
			detector: newTestDetector(t, bundle, WithDetectorPriority(DetectorSourceAccept)),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/")
				r.Header.Add("Accept-Language", "zh;q=0.1")
				r.Header.Add("Accept-Language", "ja;q=0.9")
				return r
			}(),
			want: "ja-JP",
		},
		{
			name:     "invalid explicit locale falls through to accept language",
			detector: newTestDetector(t, bundle),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/?lang=bad-locale")
				r.Header.Set("Accept-Language", "ja;q=0.9")
				return r
			}(),
			want: "ja-JP",
		},
		{
			name:     "custom priority uses header before query",
			detector: newTestDetector(t, bundle, WithDetectorPriority(DetectorSourceHeader, DetectorSourceQuery, DetectorSourceAccept)),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/?lang=zh-CN")
				r.Header.Set("X-Language", "ja")
				return r
			}(),
			want: "ja-JP",
		},
		{
			name: "unsupported accept language falls through to later header",
			detector: newTestDetector(
				t, bundle,
				WithDetectorPriority(DetectorSourceAccept, DetectorSourceHeader),
			),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/")
				r.Header.Set("Accept-Language", "de;q=0.9")
				r.Header.Set("X-Language", "ja")
				return r
			}(),
			want: "ja-JP",
		},
		{
			name: "malformed accept language falls through to later header",
			detector: newTestDetector(
				t, bundle,
				WithDetectorPriority(DetectorSourceAccept, DetectorSourceHeader),
			),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/")
				r.Header.Set("Accept-Language", "not-a-valid-header!!!")
				r.Header.Set("X-Language", "zh")
				return r
			}(),
			want: "zh-Hans",
		},
		{
			name: "supported accept language wins before later header",
			detector: newTestDetector(
				t, bundle,
				WithDetectorPriority(DetectorSourceAccept, DetectorSourceHeader),
			),
			request: func() *http.Request {
				r := httptestNewRequest(t, "GET", "/")
				r.Header.Set("Accept-Language", "zh;q=0.9")
				r.Header.Set("X-Language", "ja")
				return r
			}(),
			want: "zh-Hans",
		},
		{
			name:     "disabled query source skips query parameter",
			detector: newTestDetector(t, bundle, WithDetectorPriority(DetectorSourceAccept), WithDetectorQueryParam("")),
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

func TestDetectorDetectLocaleHandlesNilRequestInputs(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t,
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello"},
		"zh-Hans": {"hello": "你好"},
	}))

	detector := newTestDetector(t, bundle)
	assert.Equal(t, "en", detector.DetectLocale(nil))

	request := &http.Request{Header: make(http.Header)}
	request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

	assert.Equal(t, "zh-Hans", detector.DetectLocale(request))
}

func TestDetectorExplicitLocaleMatchesSupportedLocaleBeforeTranslationsLoad(t *testing.T) {
	t.Parallel()
	bundle := newTestBundle(t,
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en": {"hello": "Hello"},
	}))

	detector := newTestDetector(t, bundle)
	request := httptestNewRequest(t, "GET", "/?lang=zh-CN")

	assert.Equal(t, "zh-Hans", detector.DetectLocale(request))
}

func TestNewDetectorRejectsNilBundle(t *testing.T) {
	t.Parallel()

	detector, err := NewDetector(nil)
	require.Error(t, err)
	assert.Nil(t, detector)
	assert.Contains(t, err.Error(), "bundle")
}

func TestNewDetectorRejectsNilOption(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t, WithDefaultLocale("en"))
	var detector *Detector
	var err error
	require.NotPanics(t, func() {
		detector, err = NewDetector(bundle, nil)
	})
	require.Error(t, err)
	assert.Nil(t, detector)
	assert.Contains(t, err.Error(), "option")
}

func TestNewDetectorRejectsMixedInvalidPrioritySources(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t,
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)

	detector, err := NewDetector(
		bundle,
		WithDetectorPriority(DetectorSource("invalid"), DetectorSourceHeader),
	)
	require.Error(t, err)
	assert.Nil(t, detector)
	assert.Contains(t, err.Error(), `"invalid"`)
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

func TestNewDetectorRejectsAllInvalidPrioritySources(t *testing.T) {
	t.Parallel()

	bundle := newTestBundle(t,
		WithDefaultLocale("en"),
		WithLocales("en", "zh-Hans"),
	)

	detector, err := NewDetector(bundle, WithDetectorPriority(DetectorSource("bad")))
	require.Error(t, err)
	assert.Nil(t, detector)
	assert.Contains(t, err.Error(), `"bad"`)
}

func httptestNewRequest(tb testing.TB, method, target string) *http.Request {
	tb.Helper()

	req, err := http.NewRequest(method, target, nil)
	require.NoError(tb, err)
	return req
}

func newTestDetector(tb testing.TB, bundle *I18n, opts ...DetectorOption) *Detector {
	tb.Helper()

	detector, err := NewDetector(bundle, opts...)
	require.NoError(tb, err)
	return detector
}
