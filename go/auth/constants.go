package auth

import "time"

const (
	// TokenExpirationDuration is the default JWT token expiration (1 week).
	TokenExpirationDuration = time.Hour * 24 * 7

	// Role constants
	RoleAdmin       = "ADMIN"
	RoleAnalyst     = "ANALYST"
	RoleViewer      = "VIEWER"
	ScopeTypeAccess = "access"
	SMAPAPI         = "smap-api"
)

// Context keys for storing auth data
type PayloadCtxKey struct{}
type ScopeCtxKey struct{}
