package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/smap-hcmut/shared-libs/go/auth"
	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// Middleware is a composite authentication middleware for Gin services.
// It combines JWT auth, role-based authorization, and internal service auth
// in a single reusable struct — eliminating the need for per-service wrapper packages.
type Middleware struct {
	authMw      *auth.Middleware
	internalKey string
}

// Config holds configuration for creating a Middleware instance.
type Config struct {
	JWTManager       auth.Manager
	CookieName       string                // default: "smap_auth_token"
	ProductionDomain string                // optional, e.g. ".tantai.dev"
	InternalKey      string                // key for X-Internal-Key header validation
	BlacklistRedis   auth.BlacklistChecker // optional: Redis client for token blacklist
	IsProduction     bool                  // when true: Bearer disabled; when false: Bearer allowed for dev
	Tracer           tracing.TraceContext  // optional: defaults to NewTraceContext()
}

// New creates a composite Middleware for Gin services.
func New(cfg Config) *Middleware {
	authMw := auth.NewMiddleware(auth.MiddlewareConfig{
		Manager:          cfg.JWTManager,
		BlacklistRedis:   cfg.BlacklistRedis,
		CookieName:       cfg.CookieName,
		Tracer:           cfg.Tracer,
		IsProduction:     cfg.IsProduction,
		ProductionDomain: cfg.ProductionDomain,
	})
	return &Middleware{
		authMw:      authMw,
		internalKey: cfg.InternalKey,
	}
}

// Auth returns a Gin middleware that validates JWT tokens.
// Stores the token payload in the request context for downstream handlers.
func (m *Middleware) Auth() gin.HandlerFunc {
	return m.authMw.Authenticate()
}

// AdminOnly returns a Gin middleware that requires the ADMIN role.
// Must be chained after Auth().
func (m *Middleware) AdminOnly() gin.HandlerFunc {
	return m.authMw.RequireRole(auth.RoleAdmin)
}

// InternalAuth returns a Gin middleware that validates the X-Internal-Key header.
// Used for service-to-service routes that should not be publicly accessible.
func (m *Middleware) InternalAuth() gin.HandlerFunc {
	return auth.InternalAuth(auth.InternalAuthConfig{ExpectedKey: m.internalKey})
}
