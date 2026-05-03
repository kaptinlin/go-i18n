// Package middleware provides optional HTTP integration for go-i18n.
package middleware

import (
	"net/http"

	i18n "github.com/kaptinlin/go-i18n"
)

// Option configures the HTTP middleware.
type Option func(*config)

type config struct {
	detector *i18n.Detector
}

// WithDetector sets a custom detector.
func WithDetector(detector *i18n.Detector) Option {
	return func(cfg *config) {
		if detector == nil {
			return
		}
		cfg.detector = detector
	}
}

// HTTPMiddleware injects a request-scoped localizer into the request context.
func HTTPMiddleware(bundle *i18n.I18n, opts ...Option) func(http.Handler) http.Handler {
	cfg := config{detector: i18n.NewDetector(bundle)}
	for _, opt := range opts {
		opt(&cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			localizer := bundle.NewLocalizer(cfg.detector.DetectLocale(r))
			ctx := i18n.ContextWithLocalizer(r.Context(), localizer)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
