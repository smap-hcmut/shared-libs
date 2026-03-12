package locale

// Locale is a context key type for storing locale information with trace integration
type Locale struct{}

// LocaleManager provides locale management with trace integration
type LocaleManager interface {
	// ParseLang parses and validates a language code with trace context
	ParseLang(lang string) string
	// IsValidLang checks if a language code is supported
	IsValidLang(lang string) bool
	// GetSupportedLanguages returns all supported language codes
	GetSupportedLanguages() []string
	// GetDefaultLanguage returns the default language
	GetDefaultLanguage() string
}
