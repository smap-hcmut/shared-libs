package auth

import (
	"github.com/golang-jwt/jwt"
)

// Payload represents JWT token payload compatible with existing services.
// This matches the structure used in ingest-srv and project-srv.
type Payload struct {
	UserID   string `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	Role     string `json:"role,omitempty"`
	Type     string `json:"type,omitempty"`
	Refresh  bool   `json:"refresh,omitempty"`
	jwt.StandardClaims
}

// Scope represents user scope information compatible with existing services.
// This matches the model.Scope structure used in services.
type Scope struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// NewScope builds Scope from Payload.
// This function is compatible with existing service implementations.
func NewScope(payload Payload) Scope {
	userID := payload.UserID
	if userID == "" {
		userID = payload.Subject
	}
	return Scope{
		UserID:   userID,
		Username: payload.Username,
		Role:     payload.Role,
	}
}
