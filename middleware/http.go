// Package middleware provides optional HTTP integration for go-i18n.
package middleware

import (
	"net/http"

	i18n "github.com/kaptinlin/go-i18n"
)

// HTTPMiddleware returns middleware that injects a request-scoped localizer
// into the request context. Detector setup errors are returned immediately.
func HTTPMiddleware(
	bundle *i18n.I18n,
	opts ...i18n.DetectorOption,
) (func(http.Handler) http.Handler, error) {
	detector, err := i18n.NewDetector(bundle, opts...)
	if err != nil {
		return nil, err
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			localizer := bundle.NewLocalizer(detector.DetectLocale(r))
			ctx := i18n.ContextWithLocalizer(r.Context(), localizer)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}, nil
}
