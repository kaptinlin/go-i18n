package i18n

import "net/http"

// DetectorSource identifies a request input that can supply a locale.
type DetectorSource string

// Locale detector sources.
const (
	DetectorSourceQuery  DetectorSource = "query"
	DetectorSourceCookie DetectorSource = "cookie"
	DetectorSourceHeader DetectorSource = "header"
	DetectorSourceAccept DetectorSource = "accept-language"
)

// DetectorOption configures a [Detector].
type DetectorOption func(*Detector)

// Detector resolves the best locale for an HTTP request.
type Detector struct {
	bundle     *I18n
	priority   []DetectorSource
	queryParam string
	cookieName string
	headerName string
}

// NewDetector creates a request locale detector for the given bundle.
func NewDetector(bundle *I18n, opts ...DetectorOption) *Detector {
	d := &Detector{
		bundle:     bundle,
		priority:   []DetectorSource{DetectorSourceQuery, DetectorSourceCookie, DetectorSourceHeader, DetectorSourceAccept},
		queryParam: "lang",
		cookieName: "lang",
		headerName: "X-Language",
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// WithDetectorPriority sets the detector source priority.
func WithDetectorPriority(priority ...DetectorSource) DetectorOption {
	return func(d *Detector) {
		if len(priority) == 0 {
			return
		}

		sanitized := make([]DetectorSource, 0, len(priority))
		for _, source := range priority {
			if source.isValid() {
				sanitized = append(sanitized, source)
			}
		}
		if len(sanitized) == 0 {
			return
		}

		d.priority = sanitized
	}
}

// WithDetectorQueryParam sets the query parameter name for locale detection.
func WithDetectorQueryParam(name string) DetectorOption {
	return func(d *Detector) {
		d.queryParam = name
	}
}

// WithDetectorCookieName sets the cookie name for locale detection.
func WithDetectorCookieName(name string) DetectorOption {
	return func(d *Detector) {
		d.cookieName = name
	}
}

// WithDetectorHeaderName sets the header name for locale detection.
func WithDetectorHeaderName(name string) DetectorOption {
	return func(d *Detector) {
		d.headerName = name
	}
}

// DetectLocale returns the best matching locale for r.
func (d *Detector) DetectLocale(r *http.Request) string {
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
	if d.queryParam == "" {
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
	value := r.Header.Get("Accept-Language")
	if value == "" {
		return ""
	}
	return d.bundle.MatchAvailableLocale(value)
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
