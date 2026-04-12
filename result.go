package i18n

// TranslationSource describes how a lookup result was produced.
type TranslationSource string

const (
	TranslationSourceDirect   TranslationSource = "direct"
	TranslationSourceFallback TranslationSource = "fallback"
	TranslationSourceMissing  TranslationSource = "missing"
)

// TranslationResult holds detailed translation lookup information.
type TranslationResult struct {
	// Text is the translated message, or the key itself if not found.
	Text string

	// Locale is the BCP 47 locale tag that produced Text.
	Locale string

	// Source reports whether the result came from the requested locale,
	// the fallback chain, or runtime key fallback.
	Source TranslationSource
}
