package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/smap-hcmut/shared-libs/go/response"
)

const defaultInternalAuthHeader = "X-Internal-Key"

// InternalAuthConfig holds configuration for internal service authentication.
type InternalAuthConfig struct {
	ExpectedKey string
}

// InternalAuth validates the shared internal header used for service-to-service calls.
func InternalAuth(cfg InternalAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.ExpectedKey == "" {
			unauthorizedInternalAuth(c)
			return
		}

		if c.GetHeader(defaultInternalAuthHeader) != cfg.ExpectedKey {
			unauthorizedInternalAuth(c)
			return
		}

		c.Set("auth_type", "internal")
		c.Next()
	}
}

func unauthorizedInternalAuth(c *gin.Context) {
	response.Unauthorized(c)
	c.Abort()
}
