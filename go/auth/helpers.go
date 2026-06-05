package auth

import (
	"context"
)

// HasPermission checks if user has a specific permission with trace logging
func HasPermission(ctx context.Context, permission string) bool {
	payload, ok := GetPayloadFromContext(ctx)
	if !ok {
		return false
	}

	// For now, map permissions to roles
	// In the future, this could check against a permission database
	switch permission {
	case "campaigns:create", "campaigns:update", "campaigns:delete",
		"projects:create", "projects:update", "projects:delete",
		"datasources:create", "datasources:update", "datasources:delete",
		"targets:create", "targets:update", "targets:delete",
		"pipeline:trigger", "crisis:configure", "ontology:configure",
		"users:manage":
		return payload.Role == "ADMIN"
	case "campaigns:read", "projects:read", "datasources:read", "targets:read":
		return payload.Role == "VIEWER" || payload.Role == "ANALYST" || payload.Role == "ADMIN"
	default:
		return false
	}
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) string {
	userID, _ := GetUserIDFromContext(ctx)
	return userID
}

// GetUserRole retrieves user role from context
func GetUserRole(ctx context.Context) string {
	payload, ok := GetPayloadFromContext(ctx)
	if !ok {
		return ""
	}
	return payload.Role
}

// GetUserEmail retrieves user email from context (if available in username field)
func GetUserEmail(ctx context.Context) string {
	payload, ok := GetPayloadFromContext(ctx)
	if !ok {
		return ""
	}
	return payload.Username // In many cases, username is email
}

// IsAuthenticated checks if request is authenticated
func IsAuthenticated(ctx context.Context) bool {
	_, ok := GetPayloadFromContext(ctx)
	return ok
}

// IsAdmin checks if user has ADMIN role
func IsAdmin(ctx context.Context) bool {
	payload, ok := GetPayloadFromContext(ctx)
	if !ok {
		return false
	}
	return payload.Role == "ADMIN"
}

// IsAnalyst checks if user has ANALYST role
func IsAnalyst(ctx context.Context) bool {
	payload, ok := GetPayloadFromContext(ctx)
	if !ok {
		return false
	}
	return payload.Role == "ANALYST"
}

// IsViewer checks if user has VIEWER role
func IsViewer(ctx context.Context) bool {
	payload, ok := GetPayloadFromContext(ctx)
	if !ok {
		return false
	}
	return payload.Role == "VIEWER"
}

// CanAccessResource checks if user can access a resource
func CanAccessResource(ctx context.Context, resourceOwnerID string) bool {
	payload, ok := GetPayloadFromContext(ctx)
	if !ok {
		return false
	}

	// Admin can access all resources
	if payload.Role == "ADMIN" {
		return true
	}

	// User can access their own resources
	if payload.UserID == resourceOwnerID {
		return true
	}

	return false
}

// RequirePermission is a helper function to check permission in handlers
func RequirePermission(ctx context.Context, permission string) error {
	if !HasPermission(ctx, permission) {
		return ErrInsufficientPermissions
	}
	return nil
}

// RequireRoleFunc is a helper function to check role in handlers
func RequireRoleFunc(ctx context.Context, role string) error {
	payload, ok := GetPayloadFromContext(ctx)
	if !ok {
		return ErrMissingToken
	}

	if payload.Role != role {
		return ErrInsufficientPermissions
	}

	return nil
}

// RequireAnyRoleFunc is a helper function to check any role in handlers
func RequireAnyRoleFunc(ctx context.Context, roles ...string) error {
	payload, ok := GetPayloadFromContext(ctx)
	if !ok {
		return ErrMissingToken
	}

	for _, role := range roles {
		if payload.Role == role {
			return nil
		}
	}

	return ErrInsufficientPermissions
}
