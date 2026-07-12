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

	bundle, err := i18n.NewBundle("en",
		i18n.WithLocales("zh-Hans", "ja-JP"),
	)
	require.NoError(t, err)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello"},
		"zh-Hans": {"hello": "你好"},
		"ja-JP":   {"hello": "こんにちは"},
	}))

	localize, err := HTTPMiddleware(bundle)
	require.NoError(t, err)

	h := localize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loc, ok := i18n.LocalizerFromContext(r.Context())
		require.True(t, ok)
		_, _ = w.Write([]byte(loc.Locale() + ":" + loc.Get("hello")))
	}))

	req := httptest.NewRequest(http.MethodGet, "/?lang=ja", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, "ja-JP:こんにちは", rr.Body.String())
}

func TestHTTPMiddlewareWithDetectorOptions(t *testing.T) {
	t.Parallel()

	bundle, err := i18n.NewBundle("en",
		i18n.WithLocales("zh-Hans", "ja-JP"),
	)
	require.NoError(t, err)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":      {"hello": "Hello"},
		"zh-Hans": {"hello": "你好"},
		"ja-JP":   {"hello": "こんにちは"},
	}))

	localize, err := HTTPMiddleware(
		bundle,
		i18n.WithDetectorPriority(i18n.DetectorSourceHeader),
		i18n.WithDetectorHeaderName("X-Locale"),
	)
	require.NoError(t, err)

	h := localize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestHTTPMiddlewareRejectsInvalidDetectorSetup(t *testing.T) {
	t.Parallel()

	bundle, err := i18n.NewBundle("en")
	require.NoError(t, err)

	tests := []struct {
		name   string
		bundle *i18n.I18n
		opts   []i18n.DetectorOption
		want   string
	}{
		{name: "nil bundle", want: "bundle"},
		{
			name:   "unknown priority source",
			bundle: bundle,
			opts:   []i18n.DetectorOption{i18n.WithDetectorPriority(i18n.DetectorSource("bad"))},
			want:   `"bad"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			localize, err := HTTPMiddleware(tt.bundle, tt.opts...)
			require.Error(t, err)
			assert.Nil(t, localize)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}
