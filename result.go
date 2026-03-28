package i18n

// TranslationResult holds detailed translation lookup information.
// Fields map directly to Google AIP-193 LocalizedMessage.
type TranslationResult struct {
	// Text is the translated message, or the key itself if not found.
	// Maps to AIP-193 LocalizedMessage.message — always populated.
	Text string

	// Locale is the BCP 47 locale tag that provided the translation.
	// Maps to AIP-193 LocalizedMessage.locale — always populated.
	// On a direct hit, this is the requested locale.
	// On a fallback hit, this is the fallback locale.
	// On a miss, this is the default locale.
	Locale string

	// Found reports whether the key existed in loaded translations.
	Found bool
}
