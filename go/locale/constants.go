package locale

// Supported language constants
const (
	EN = "en" // English
	VI = "vi" // Vietnamese
	JA = "ja" // Japanese
)

// LangList contains all supported language codes
var LangList = []string{EN, VI, JA}

// DefaultLang is the default language used when no valid locale is provided
var DefaultLang = EN
