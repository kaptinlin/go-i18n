package i18n

// TranslationSource describes how a lookup result was produced.
type TranslationSource string

const (
	// TranslationSourceDirect indicates that the requested locale supplied the translation.
	TranslationSourceDirect TranslationSource = "direct"
	// TranslationSourceFallback indicates that a fallback locale supplied the translation.
	TranslationSourceFallback TranslationSource = "fallback"
	// TranslationSourceMissing indicates that no loaded translation was found and the key was returned.
	TranslationSourceMissing TranslationSource = "missing"
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
