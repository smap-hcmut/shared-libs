package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// Middleware handles JWT authentication with trace integration
type Middleware struct {
	manager        Manager
	blacklistRedis BlacklistChecker
	cookieName     string
	tracer         tracing.TraceContext
}

// BlacklistChecker interface for checking if token is blacklisted
type BlacklistChecker interface {
	Exists(ctx context.Context, key string) (bool, error)
}

// MiddlewareConfig holds configuration for middleware
type MiddlewareConfig struct {
	Manager        Manager
	BlacklistRedis BlacklistChecker // Optional
	CookieName     string
	Tracer         tracing.TraceContext // Optional, will create default if nil
}

// NewMiddleware creates a new authentication middleware with trace integration
func NewMiddleware(cfg MiddlewareConfig) *Middleware {
	if cfg.CookieName == "" {
		cfg.CookieName = "smap_auth_token"
	}
	if cfg.Tracer == nil {
		cfg.Tracer = tracing.NewTraceContext()
	}

	return &Middleware{
		manager:        cfg.Manager,
		blacklistRedis: cfg.BlacklistRedis,
		cookieName:     cfg.CookieName,
		tracer:         cfg.Tracer,
	}
}

// Authenticate is a Gin middleware that verifies JWT tokens with trace integration
func (m *Middleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Extract token from request
		tokenString, err := m.extractToken(c)
		if err != nil {
			m.respondWithError(c, http.StatusUnauthorized, "MISSING_TOKEN",
				"Authentication token is required", err.Error())
			return
		}

		// Verify token with trace integration
		payload, enhancedCtx, err := m.manager.VerifyWithTrace(ctx, tokenString)
		if err != nil {
			m.respondWithError(c, http.StatusUnauthorized, "INVALID_TOKEN",
				"Invalid or expired authentication token", err.Error())
			return
		}

		// Check if token is blacklisted
		if m.blacklistRedis != nil {
			blacklisted, err := m.isBlacklisted(enhancedCtx, payload.Id)
			if err != nil {
				m.respondWithError(c, http.StatusInternalServerError, "BLACKLIST_CHECK_FAILED",
					"Failed to verify token status", err.Error())
				return
			}

			if blacklisted {
				m.respondWithError(c, http.StatusUnauthorized, "TOKEN_REVOKED",
					"This token has been revoked", "")
				return
			}
		}

		// Update request context with enhanced context (includes trace_id and payload)
		c.Request = c.Request.WithContext(enhancedCtx)

		c.Next()
	}
}

// extractToken extracts JWT token from Authorization header or cookie
func (m *Middleware) extractToken(c *gin.Context) (string, error) {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Expected format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1], nil
		}
	}

	// Try cookie
	token, err := c.Cookie(m.cookieName)
	if err == nil && token != "" {
		return token, nil
	}

	return "", ErrMissingToken
}

// isBlacklisted checks if token ID is in blacklist
func (m *Middleware) isBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := "blacklist:token:" + tokenID
	return m.blacklistRedis.Exists(ctx, key)
}

// respondWithError sends error response with trace_id integration
func (m *Middleware) respondWithError(c *gin.Context, status int, code, message, details string) {
	ctx := c.Request.Context()
	traceID := m.tracer.GetTraceID(ctx)

	response := gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	}

	// Add trace_id to error response if available
	if traceID != "" {
		response["trace_id"] = traceID
	}

	// Add details if provided
	if details != "" {
		response["error"].(gin.H)["details"] = details
	}

	c.JSON(status, response)
	c.Abort()
}

// RequireRole returns a middleware that requires a specific role
func (m *Middleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		payload, ok := GetPayloadFromContext(ctx)
		if !ok {
			m.respondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED",
				"Authentication required", "")
			return
		}

		if payload.Role != role {
			m.respondWithError(c, http.StatusForbidden, "INSUFFICIENT_PERMISSIONS",
				"You do not have permission to access this resource",
				"required_role: "+role+", your_role: "+payload.Role)
			return
		}

		c.Next()
	}
}

// RequireAnyRole returns a middleware that requires any of the specified roles
func (m *Middleware) RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		payload, ok := GetPayloadFromContext(ctx)
		if !ok {
			m.respondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED",
				"Authentication required", "")
			return
		}

		hasRole := false
		for _, role := range roles {
			if payload.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			m.respondWithError(c, http.StatusForbidden, "INSUFFICIENT_PERMISSIONS",
				"You do not have permission to access this resource",
				"required_roles: "+strings.Join(roles, ", ")+", your_role: "+payload.Role)
			return
		}

		c.Next()
	}
}
