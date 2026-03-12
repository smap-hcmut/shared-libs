# Locale Package

The locale package provides internationalization (i18n) and locale management with distributed tracing integration for SMAP services.

## Features

- **Multi-language Support**: English, Vietnamese, and Japanese
- **Trace Integration**: Automatic trace_id propagation in locale operations
- **Context Management**: Store and retrieve locale from context
- **Validation**: Language code validation and normalization
- **Backward Compatibility**: Drop-in replacement for existing locale packages
- **Flexible Parsing**: Case-insensitive language parsing with aliases

## Supported Languages

- **English** (`en`): Default language
- **Vietnamese** (`vi`): Aliases: "vietnamese", "việt nam"
- **Japanese** (`ja`): Aliases: "japanese"

## Usage

### Basic Usage (Backward Compatible)

```go
import "github.com/smap-hcmut/shared-libs/go/locale"

// Parse language codes
lang := locale.ParseLang("EN")        // Returns "en"
lang = locale.ParseLang("Vietnamese") // Returns "vi"
lang = locale.ParseLang("invalid")    // Returns "en" (default)

// Validate language codes
valid := locale.IsValidLang("vi")     // Returns true
valid = locale.IsValidLang("invalid") // Returns false

// Context management
ctx = locale.SetLocaleToContext(ctx, "vi")
lang := locale.GetLang(ctx)           // Returns "vi"
lang, ok := locale.GetLocaleFromContext(ctx)
```

### Advanced Usage with Trace Integration

```go
import (
    "github.com/smap-hcmut/shared-libs/go/locale"
    "github.com/smap-hcmut/shared-libs/go/tracing"
)

// Create manager with trace integration
manager := locale.NewManager()

// Or with custom tracer
tracer := tracing.NewTraceContext()
manager := locale.NewManagerWithTracer(tracer)

// Use manager methods
lang := manager.ParseLang("vietnamese")
valid := manager.IsValidLang("ja")
supported := manager.GetSupportedLanguages()
defaultLang := manager.GetDefaultLanguage()

// Trace-aware functions
lang := locale.ParseLangWithTrace(ctx, "EN")
valid := locale.IsValidLangWithTrace(ctx, "vi")
ctx = locale.SetLocaleToContextWithTrace(ctx, "ja", tracer)
lang, ok := locale.GetLocaleFromContextWithTrace(ctx)
```

### HTTP Middleware Integration

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/smap-hcmut/shared-libs/go/locale"
)

func LocaleMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get language from header, query param, or cookie
        lang := c.GetHeader("Accept-Language")
        if lang == "" {
            lang = c.Query("lang")
        }
        if lang == "" {
            lang, _ = c.Cookie("lang")
        }
        
        // Parse and set in context
        parsedLang := locale.ParseLang(lang)
        ctx := locale.SetLocaleToContext(c.Request.Context(), parsedLang)
        c.Request = c.Request.WithContext(ctx)
        
        c.Next()
    }
}

// Usage in handlers
func MyHandler(c *gin.Context) {
    lang := locale.GetLang(c.Request.Context())
    // Use lang for localized responses
}
```

### Service-to-Service Communication

```go
// Set locale header for outgoing requests
lang := locale.GetLang(ctx)
req.Header.Set("Accept-Language", lang)

// Parse locale from incoming requests
lang := locale.ParseLang(req.Header.Get("Accept-Language"))
ctx = locale.SetLocaleToContext(ctx, lang)
```

## API Reference

### Functions

#### Language Parsing
- `ParseLang(lang string) string`: Parse and validate language code
- `ParseLangWithTrace(ctx context.Context, lang string) string`: Parse with trace context

#### Validation
- `IsValidLang(lang string) bool`: Check if language is supported
- `IsValidLangWithTrace(ctx context.Context, lang string) bool`: Validate with trace context

#### Context Management
- `SetLocaleToContext(ctx context.Context, lang string) context.Context`: Set locale in context
- `SetLocaleToContextWithTrace(ctx, lang, tracer)`: Set with explicit tracer
- `GetLocaleFromContext(ctx context.Context) (string, bool)`: Get locale from context
- `GetLocaleFromContextWithTrace(ctx)`: Get with trace context
- `GetLang(ctx context.Context) string`: Get locale or default
- `GetLangWithTrace(ctx context.Context) string`: Get with trace context

#### Manager Interface
- `NewManager() LocaleManager`: Create manager with default tracer
- `NewManagerWithTracer(tracer) LocaleManager`: Create with custom tracer

### Manager Methods
- `ParseLang(lang string) string`: Parse language code
- `IsValidLang(lang string) bool`: Validate language code
- `GetSupportedLanguages() []string`: Get all supported languages
- `GetDefaultLanguage() string`: Get default language

## Constants

### Languages
- `EN`: English language code ("en")
- `VI`: Vietnamese language code ("vi")
- `JA`: Japanese language code ("ja")

### Variables
- `LangList`: Slice of all supported language codes
- `DefaultLang`: Default language (English)

### Errors
- `ErrLocaleNotFound`: Returned when locale is not supported

## Migration Guide

### From Local Locale Package

1. Update imports:
```go
// Before
import "your-service/pkg/locale"

// After
import "github.com/smap-hcmut/shared-libs/go/locale"
```

2. No code changes needed for basic usage
3. Optional: Add trace integration for enhanced debugging

### Language Code Mapping

The package automatically handles various language formats:
- Case-insensitive: "EN", "en", "En" → "en"
- Full names: "English", "Vietnamese" → "en", "vi"
- Native names: "việt nam" → "vi"
- Invalid codes → Default language ("en")

### Trace Integration Benefits

- **Request Tracking**: Follow locale operations across services
- **Localization Debugging**: Track language selection and propagation
- **Performance Monitoring**: Measure locale processing overhead
- **User Experience**: Better debugging of localization issues

## Best Practices

1. **Always validate**: Use `ParseLang()` instead of direct assignment
2. **Context propagation**: Maintain locale in request context
3. **Service boundaries**: Pass locale in headers between services
4. **Fallback handling**: Always provide default language fallback
5. **Trace integration**: Use trace-aware functions for better debugging