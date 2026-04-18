package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	i18n "github.com/kaptinlin/go-i18n"
)

func TestHTTPMiddleware(t *testing.T) {
	t.Parallel()

	bundle := i18n.NewBundle(
		i18n.WithDefaultLocale("en"),
		i18n.WithLocales("en", "zh-Hans", "ja-JP"),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello"},
		"zh-Hans": {"hello": "你好"},
		"ja-JP":   {"hello": "こんにちは"},
	}))

	h := HTTPMiddleware(bundle)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loc, ok := i18n.LocalizerFromContext(r.Context())
		require.True(t, ok)
		_, _ = w.Write([]byte(loc.Locale() + ":" + loc.Get("hello")))
	}))

	req := httptest.NewRequest(http.MethodGet, "/?lang=ja", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, "ja-JP:こんにちは", rr.Body.String())
}

func TestHTTPMiddlewareWithCustomDetector(t *testing.T) {
	t.Parallel()

	bundle := i18n.NewBundle(
		i18n.WithDefaultLocale("en"),
		i18n.WithLocales("en", "zh-Hans", "ja-JP"),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello"},
		"zh-Hans": {"hello": "你好"},
		"ja-JP":   {"hello": "こんにちは"},
	}))

	detector := i18n.NewDetector(bundle, i18n.WithDetectorPriority(i18n.DetectorSourceHeader), i18n.WithDetectorHeaderName("X-Locale"))
	h := HTTPMiddleware(bundle, WithDetector(detector))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loc, ok := i18n.LocalizerFromContext(r.Context())
		require.True(t, ok)
		_, _ = w.Write([]byte(loc.Locale()))
	}))

	req := httptest.NewRequest(http.MethodGet, "/?lang=zh-CN", nil)
	req.Header.Set("X-Locale", "ja")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, "ja-JP", rr.Body.String())
}

func TestHTTPMiddlewareWithNilDetectorFallsBackToDefault(t *testing.T) {
	t.Parallel()

	bundle := i18n.NewBundle(
		i18n.WithDefaultLocale("en"),
		i18n.WithLocales("en", "zh-Hans", "ja-JP"),
	)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello"},
		"zh-Hans": {"hello": "你好"},
		"ja-JP":   {"hello": "こんにちは"},
	}))

	h := HTTPMiddleware(bundle, WithDetector(nil))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loc, ok := i18n.LocalizerFromContext(r.Context())
		require.True(t, ok)
		_, _ = w.Write([]byte(loc.Locale()))
	}))

	req := httptest.NewRequest(http.MethodGet, "/?lang=ja", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, "ja-JP", rr.Body.String())
}
