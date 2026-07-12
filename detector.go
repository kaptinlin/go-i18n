package i18n

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
)

// DetectorSource identifies a request input that can supply a locale.
type DetectorSource string

const (
	// DetectorSourceQuery reads the locale from the request query string.
	DetectorSourceQuery DetectorSource = "query"
	// DetectorSourceCookie reads the locale from a request cookie.
	DetectorSourceCookie DetectorSource = "cookie"
	// DetectorSourceHeader reads the locale from a request header.
	DetectorSourceHeader DetectorSource = "header"
	// DetectorSourceAccept reads the locale from the Accept-Language header.
	DetectorSourceAccept DetectorSource = "accept-language"
)

// DetectorOption configures a [Detector] during [NewDetector].
type DetectorOption func(*Detector) error

// Detector resolves the best locale for an HTTP request.
type Detector struct {
	bundle     *I18n
	priority   []DetectorSource
	queryParam string
	cookieName string
	headerName string
}

// NewDetector creates a request locale detector for the given bundle. It
// returns an error for a nil bundle, a nil option, or an unknown priority
// source.
func NewDetector(bundle *I18n, opts ...DetectorOption) (*Detector, error) {
	if bundle == nil {
		return nil, errors.New("detector bundle is nil")
	}

	d := &Detector{
		bundle:     bundle,
		priority:   []DetectorSource{DetectorSourceQuery, DetectorSourceCookie, DetectorSourceHeader, DetectorSourceAccept},
		queryParam: "lang",
		cookieName: "lang",
		headerName: "X-Language",
	}
	for _, opt := range opts {
		if opt == nil {
			return nil, errors.New("detector option is nil")
		}
		if err := opt(d); err != nil {
			return nil, err
		}
	}
	return d, nil
}

// WithDetectorPriority sets the detector source priority. An empty priority
// keeps the default order. NewDetector rejects unknown sources.
func WithDetectorPriority(priority ...DetectorSource) DetectorOption {
	return func(d *Detector) error {
		if len(priority) == 0 {
			return nil
		}

		for _, source := range priority {
			if !source.isValid() {
				return fmt.Errorf("detector priority source %q is invalid", source)
			}
		}

		d.priority = slices.Clone(priority)
		return nil
	}
}

// WithDetectorQueryParam sets the query parameter name for locale detection.
func WithDetectorQueryParam(name string) DetectorOption {
	return func(d *Detector) error {
		d.queryParam = name
		return nil
	}
}

// WithDetectorCookieName sets the cookie name for locale detection.
func WithDetectorCookieName(name string) DetectorOption {
	return func(d *Detector) error {
		d.cookieName = name
		return nil
	}
}

// WithDetectorHeaderName sets the header name for locale detection.
func WithDetectorHeaderName(name string) DetectorOption {
	return func(d *Detector) error {
		d.headerName = name
		return nil
	}
}

// DetectLocale returns the best matching locale for r.
func (d *Detector) DetectLocale(r *http.Request) string {
	if r == nil {
		return d.bundle.defaultLocale
	}

	for _, source := range d.priority {
		switch source {
		case DetectorSourceQuery:
			if locale := d.detectQuery(r); locale != "" {
				return locale
			}
		case DetectorSourceCookie:
			if locale := d.detectCookie(r); locale != "" {
				return locale
			}
		case DetectorSourceHeader:
			if locale := d.detectHeader(r); locale != "" {
				return locale
			}
		case DetectorSourceAccept:
			if locale := d.detectAcceptLanguage(r); locale != "" {
				return locale
			}
		}
	}
	return d.bundle.defaultLocale
}

func (d *Detector) detectQuery(r *http.Request) string {
	if d.queryParam == "" || r.URL == nil {
		return ""
	}
	return d.resolveExplicitLocale(r.URL.Query().Get(d.queryParam))
}

func (d *Detector) detectCookie(r *http.Request) string {
	if d.cookieName == "" {
		return ""
	}
	cookie, err := r.Cookie(d.cookieName)
	if err != nil {
		return ""
	}
	return d.resolveExplicitLocale(cookie.Value)
}

func (d *Detector) detectHeader(r *http.Request) string {
	if d.headerName == "" {
		return ""
	}
	return d.resolveExplicitLocale(r.Header.Get(d.headerName))
}

func (d *Detector) detectAcceptLanguage(r *http.Request) string {
	values := r.Header.Values("Accept-Language")
	if len(values) == 0 {
		return ""
	}
	locale, _ := d.bundle.matchAvailableLocale(values...)
	return locale
}

func (d *Detector) resolveExplicitLocale(locale string) string {
	if locale == "" {
		return ""
	}
	matched, ok := d.bundle.resolveLocalizedLocale(locale)
	if !ok {
		return ""
	}
	return matched
}

func (s DetectorSource) isValid() bool {
	switch s {
	case DetectorSourceQuery, DetectorSourceCookie, DetectorSourceHeader, DetectorSourceAccept:
		return true
	default:
		return false
	}
}
