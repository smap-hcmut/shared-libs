package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CookieConfig holds HTTP cookie configuration.
type CookieConfig struct {
	Name     string
	MaxAge   int
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

// GetDynamicCookieConfig returns cookie configuration based on request Origin header.
// For localhost origins: returns SameSite=None (for cross-origin from localhost UI to production API)
// For other origins: returns SameSite=Lax with the specified productionDomain
func GetDynamicCookieConfig(r *http.Request, productionDomain string) CookieConfig {
	origin := r.Header.Get("Origin")

	config := CookieConfig{
		Name:   "smap_auth_token",
		MaxAge: 28800, // 8 hours
		Secure: true,
	}

	// Check if request is from localhost (development/local testing)
	if isLocalhost(origin) {
		// Cross-origin from localhost - need SameSite=None for cookie to be sent
		config.SameSite = http.SameSiteNoneMode
		config.Secure = true // Required for SameSite=None
		// Don't set Domain for cross-origin requests
	} else {
		// Same-origin or production - use Lax mode
		config.SameSite = http.SameSiteLaxMode
		if productionDomain != "" {
			config.Domain = productionDomain
		}
	}

	return config
}

// isLocalhost checks if origin is localhost.
func isLocalhost(origin string) bool {
	return strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "https://localhost")
}

// SetAuthCookie sets an HTTP-only authentication cookie with dynamic configuration based on Origin header.
// productionDomain should be something like ".tantai.dev" for production environments.
func SetAuthCookie(w http.ResponseWriter, r *http.Request, token string, productionDomain string) {
	config := GetDynamicCookieConfig(r, productionDomain)

	cookie := &http.Cookie{
		Name:     config.Name,
		Value:    token,
		HttpOnly: true, // Prevents JavaScript access for security
		Path:     "/",
		MaxAge:   config.MaxAge,
		Domain:   config.Domain,
		SameSite: config.SameSite,
		Secure:   config.Secure,
	}

	http.SetCookie(w, cookie)
}

// GinSetAuthCookie sets an HTTP-only authentication cookie in a Gin context with dynamic configuration.
// This is a convenience wrapper around SetAuthCookie for Gin applications.
func GinSetAuthCookie(c *gin.Context, token string, productionDomain string) {
	SetAuthCookie(c.Writer, c.Request, token, productionDomain)
}
