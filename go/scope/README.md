# Scope Package

The scope package provides JWT token management and user scope handling with distributed tracing integration for SMAP services.

## Features

- **JWT Token Management**: Create and verify JWT tokens with standard claims
- **Scope Management**: Handle user scope information (UserID, Username, Role)
- **Trace Integration**: Automatic trace_id propagation in scope operations
- **Context Management**: Store and retrieve scope/payload from context
- **Role-based Access**: Support for ADMIN, ANALYST, and VIEWER roles
- **Backward Compatibility**: Drop-in replacement for existing scope packages

## Usage

### Basic Usage (Backward Compatible)

```go
import "github.com/smap-hcmut/shared-libs/go/scope"

// Create manager
manager := scope.New("your-secret-key")

// Create token
payload := scope.Payload{
    UserID:   "user123",
    Username: "john.doe",
    Role:     scope.RoleAdmin,
    Type:     scope.ScopeTypeAccess,
}
token, err := manager.CreateToken(payload)

// Verify token
verifiedPayload, err := manager.Verify(token)

// Create scope from payload
userScope := scope.NewScope(verifiedPayload)
```

### Advanced Usage with Trace Integration

```go
import (
    "github.com/smap-hcmut/shared-libs/go/scope"
    "github.com/smap-hcmut/shared-libs/go/tracing"
)

// Create manager with custom tracer
tracer := tracing.NewTraceContext()
manager := scope.NewWithTracer("your-secret-key", tracer)

// All operations with trace context
token, err := manager.CreateTokenWithTrace(ctx, payload)
verifiedPayload, err := manager.VerifyWithTrace(ctx, token)
userScope := scope.NewScopeWithTrace(ctx, verifiedPayload)
```

### Context Management

```go
// Store payload in context
ctx = scope.SetPayloadToContext(ctx, payload)

// Retrieve payload from context
payload, ok := scope.GetPayloadFromContext(ctx)

// Get specific fields
userID, ok := scope.GetUserIDFromContext(ctx)
username, ok := scope.GetUsernameFromContext(ctx)

// Store scope in context
ctx = scope.SetScopeToContext(ctx, userScope)

// Retrieve scope from context
userScope, ok := scope.GetScopeFromContext(ctx)
```

### Scope Header Management

```go
// Create scope header for service-to-service communication
header, err := scope.CreateScopeHeader(userScope)

// Parse scope header
parsedScope, err := scope.ParseScopeHeader(header)

// With trace context
header, err := scope.CreateScopeHeaderWithTrace(ctx, userScope)
parsedScope, err := scope.ParseScopeHeaderWithTrace(ctx, header)
```

### Role-based Access Control

```go
// Check user roles
if userScope.IsAdmin() {
    // Admin-only operations
}

if userScope.IsAnalyst() {
    // Analyst operations
}

if userScope.IsViewer() {
    // Viewer operations
}
```

## Types

### Payload
JWT token claims structure:
```go
type Payload struct {
    jwt.StandardClaims
    UserID   string `json:"sub"`      // Subject (user ID)
    Username string `json:"username"` // Username
    Role     string `json:"role"`     // User role
    Type     string `json:"type"`     // Token type
    Refresh  bool   `json:"refresh"`  // Refresh token flag
}
```

### Scope
User scope information:
```go
type Scope struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    JTI      string `json:"jti"`
}
```

## Constants

### Roles
- `RoleAdmin`: Full system access
- `RoleAnalyst`: Analysis and reporting access
- `RoleViewer`: Read-only access

### Token Types
- `ScopeTypeAccess`: Access token type
- `SMAPAPI`: SMAP API identifier

### Configuration
- `TokenExpirationDuration`: Default token expiration (1 week)

## Migration Guide

### From Local Scope Package

1. Update imports:
```go
// Before
import "your-service/pkg/scope"

// After
import "github.com/smap-hcmut/shared-libs/go/scope"
```

2. Update model imports (if using internal models):
```go
// Before
import "your-service/internal/model"
scope := model.Scope{...}

// After
import "github.com/smap-hcmut/shared-libs/go/scope"
userScope := scope.Scope{...}
```

3. No other code changes needed for basic usage
4. Optional: Add trace integration for enhanced debugging

### Trace Integration Benefits

- **Token Tracking**: Follow JWT operations across service boundaries
- **Scope Propagation**: Maintain trace context in scope operations
- **Security Auditing**: Enhanced logging for authentication events
- **Performance Monitoring**: Measure JWT verification latency
- **Debugging**: Easier troubleshooting with trace context

## Error Handling

The package provides structured error handling:

- `ErrInvalidToken`: Returned for invalid, expired, or malformed tokens
- All errors include trace context when available
- Graceful degradation when trace context is missing