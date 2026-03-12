package locale

import (
	"context"
	"strings"

	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// Manager implementation with trace integration
type manager struct {
	tracer tracing.TraceContext
}

// NewManager creates a new locale manager with trace integration
func NewManager() LocaleManager {
	return &manager{
		tracer: tracing.NewTraceContext(),
	}
}

// NewManagerWithTracer creates a new locale manager with custom tracer
func NewManagerWithTracer(tracer tracing.TraceContext) LocaleManager {
	if tracer == nil {
		tracer = tracing.NewTraceContext()
	}
	return &manager{
		tracer: tracer,
	}
}

// ParseLang parses and validates a language code with trace integration
func (m *manager) ParseLang(lang string) string {
	return ParseLang(lang)
}

// IsValidLang checks if a language code is supported
func (m *manager) IsValidLang(lang string) bool {
	return IsValidLang(lang)
}

// GetSupportedLanguages returns all supported language codes
func (m *manager) GetSupportedLanguages() []string {
	return LangList
}

// GetDefaultLanguage returns the default language
func (m *manager) GetDefaultLanguage() string {
	return DefaultLang
}

// ParseLang parses and validates a language code
// Returns the default language if the provided code is not supported
// Input is case-insensitive and trimmed of whitespace
func ParseLang(lang string) string {
	lang = strings.TrimSpace(strings.ToLower(lang))

	switch lang {
	case EN, "english":
		return EN
	case VI, "vietnamese", "việt nam":
		return VI
	case JA, "japanese":
		return JA
	default:
		return DefaultLang
	}
}

// ParseLangWithTrace parses language with trace context
func ParseLangWithTrace(ctx context.Context, lang string) string {
	// Could add trace logging here if needed
	return ParseLang(lang)
}

// IsValidLang checks if a language code is supported
func IsValidLang(lang string) bool {
	lang = strings.TrimSpace(strings.ToLower(lang))
	for _, supported := range LangList {
		if lang == supported {
			return true
		}
	}
	return false
}

// IsValidLangWithTrace checks language validity with trace context
func IsValidLangWithTrace(ctx context.Context, lang string) bool {
	// Could add trace logging here if needed
	return IsValidLang(lang)
}

// GetLang retrieves the locale from context, returning the default if not found
func GetLang(ctx context.Context) string {
	lang, ok := GetLocaleFromContext(ctx)
	if !ok {
		return DefaultLang
	}
	return lang
}

// GetLangWithTrace retrieves locale with trace context
func GetLangWithTrace(ctx context.Context) string {
	// Could add trace logging here if needed
	return GetLang(ctx)
}

// SetLocaleToContext sets the locale in the context with trace propagation
func SetLocaleToContext(ctx context.Context, lang string) context.Context {
	// Validate before setting
	if !IsValidLang(lang) {
		lang = DefaultLang
	}

	// Ensure trace context is preserved
	tracer := tracing.NewTraceContext()
	if traceID := tracer.GetTraceID(ctx); traceID != "" {
		ctx = tracer.SetTraceID(ctx, traceID)
	}

	return context.WithValue(ctx, Locale{}, lang)
}

// SetLocaleToContextWithTrace sets locale with explicit trace context
func SetLocaleToContextWithTrace(ctx context.Context, lang string, tracer tracing.TraceContext) context.Context {
	// Validate before setting
	if !IsValidLang(lang) {
		lang = DefaultLang
	}

	// Ensure trace context is preserved
	if tracer != nil {
		if traceID := tracer.GetTraceID(ctx); traceID != "" {
			ctx = tracer.SetTraceID(ctx, traceID)
		}
	}

	return context.WithValue(ctx, Locale{}, lang)
}

// GetLocaleFromContext retrieves the locale from context
// Returns the locale and true if found, empty string and false otherwise
func GetLocaleFromContext(ctx context.Context) (string, bool) {
	lang, ok := ctx.Value(Locale{}).(string)
	if !ok || lang == "" {
		return "", false
	}
	return lang, true
}

// GetLocaleFromContextWithTrace retrieves locale with trace context
func GetLocaleFromContextWithTrace(ctx context.Context) (string, bool) {
	// Could add trace logging here if needed
	return GetLocaleFromContext(ctx)
}
