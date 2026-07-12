package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	i18n "github.com/kaptinlin/go-i18n"
	"github.com/kaptinlin/go-i18n/middleware"
)

func TestHTTPMiddlewareAcceptsDetectorOptions(t *testing.T) {
	t.Parallel()

	bundle, err := i18n.NewBundle("en",
		i18n.WithLocales("ja-JP"),
	)
	require.NoError(t, err)
	require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
		"en":    {"hello": "Hello"},
		"ja-JP": {"hello": "こんにちは"},
	}))

	localize, err := middleware.HTTPMiddleware(
		bundle,
		i18n.WithDetectorPriority(i18n.DetectorSourceHeader),
		i18n.WithDetectorHeaderName("X-Locale"),
	)
	require.NoError(t, err)

	handler := localize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		localizer, ok := i18n.LocalizerFromContext(r.Context())
		require.True(t, ok)
		_, _ = w.Write([]byte(localizer.Locale() + ":" + localizer.Get("hello")))
	}))

	request := httptest.NewRequest(http.MethodGet, "/?lang=en", nil)
	request.Header.Set("X-Locale", "ja")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	assert.Equal(t, "ja-JP:こんにちは", response.Body.String())
}
