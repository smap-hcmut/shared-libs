package auth

import "time"

const (
	// TokenExpirationDuration is the default JWT token expiration (1 week).
	TokenExpirationDuration = time.Hour * 24 * 7
)

// Context keys for storing auth data
type PayloadCtxKey struct{}
type ScopeCtxKey struct{}
