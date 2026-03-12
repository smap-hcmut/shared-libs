package scope

import (
	"context"

	"github.com/golang-jwt/jwt"
)

// Payload represents the JWT token claims with trace integration
type Payload struct {
	jwt.StandardClaims
	UserID   string `json:"sub"`      // Subject (user ID)
	Username string `json:"username"` // Username
	Role     string `json:"role"`     // User role (ADMIN, ANALYST, VIEWER)
	Type     string `json:"type"`     // Token type (e.g., "access", "refresh")
	Refresh  bool   `json:"refresh"`  // Whether this is a refresh token
}

// Scope represents user scope information with trace integration
type Scope struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"` // ADMIN, ANALYST, or VIEWER
	JTI      string `json:"jti"`
}

// Role constants
const (
	RoleAdmin   = "ADMIN"
	RoleAnalyst = "ANALYST"
	RoleViewer  = "VIEWER"
)

// Scope type constants
const (
	ScopeTypeAccess = "access"
	SMAPAPI         = "smap-api"
)

// IsAdmin checks if the scope has admin role
func (s Scope) IsAdmin() bool {
	return s.Role == RoleAdmin
}

// IsAnalyst checks if the scope has analyst role
func (s Scope) IsAnalyst() bool {
	return s.Role == RoleAnalyst
}

// IsViewer checks if the scope has viewer role
func (s Scope) IsViewer() bool {
	return s.Role == RoleViewer
}

// Manager defines the interface for JWT/scope token management with trace integration
type Manager interface {
	// Verify verifies a JWT token and returns the payload if valid
	Verify(token string) (Payload, error)
	// VerifyWithTrace verifies a JWT token with trace context
	VerifyWithTrace(ctx context.Context, token string) (Payload, error)
	// CreateToken creates a new JWT token with the provided payload
	CreateToken(payload Payload) (string, error)
	// CreateTokenWithTrace creates a new JWT token with trace context
	CreateTokenWithTrace(ctx context.Context, payload Payload) (string, error)
	// VerifyScope parses and verifies scope header
	VerifyScope(scopeHeader string) (Scope, error)
	// VerifyScopeWithTrace parses and verifies scope header with trace context
	VerifyScopeWithTrace(ctx context.Context, scopeHeader string) (Scope, error)
}

// Context key types for payload and scope with trace integration
type (
	PayloadCtxKey      struct{}
	ScopeCtxKey        struct{}
	ThirdPartyScopeKey struct{}
	SessionUserCtxKey  struct{}
)
