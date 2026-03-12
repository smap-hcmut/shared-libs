package scope

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// NewScope builds Scope from Payload with trace integration
func NewScope(payload Payload) Scope {
	userID := payload.UserID
	if userID == "" {
		userID = payload.Subject
	}
	return Scope{
		UserID:   userID,
		Username: payload.Username,
		Role:     payload.Role,
		JTI:      payload.Id,
	}
}

// NewScopeWithTrace builds Scope from Payload with trace context
func NewScopeWithTrace(ctx context.Context, payload Payload) Scope {
	// Could add trace logging here if needed
	return NewScope(payload)
}

// CreateScopeHeader encodes scope as base64 JSON header value
func CreateScopeHeader(scope Scope) (string, error) {
	jsonData, err := json.Marshal(scope)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(jsonData), nil
}

// CreateScopeHeaderWithTrace encodes scope with trace context
func CreateScopeHeaderWithTrace(ctx context.Context, scope Scope) (string, error) {
	// Could add trace logging here if needed
	return CreateScopeHeader(scope)
}

// ParseScopeHeader decodes scope from base64 JSON header value
func ParseScopeHeader(scopeHeader string) (Scope, error) {
	jsonData, err := base64.StdEncoding.DecodeString(scopeHeader)
	if err != nil {
		return Scope{}, err
	}
	var scope Scope
	if err := json.Unmarshal(jsonData, &scope); err != nil {
		return Scope{}, err
	}
	return scope, nil
}

// ParseScopeHeaderWithTrace decodes scope with trace context
func ParseScopeHeaderWithTrace(ctx context.Context, scopeHeader string) (Scope, error) {
	// Could add trace logging here if needed
	return ParseScopeHeader(scopeHeader)
}

// Context management functions with trace integration

// SetPayloadToContext attaches Payload to context with trace propagation
func SetPayloadToContext(ctx context.Context, payload Payload) context.Context {
	// Ensure trace context is preserved
	tracer := tracing.NewTraceContext()
	if traceID := tracer.GetTraceID(ctx); traceID != "" {
		ctx = tracer.WithTraceID(ctx, traceID)
	}
	return context.WithValue(ctx, PayloadCtxKey{}, payload)
}

// GetPayloadFromContext returns Payload from context
func GetPayloadFromContext(ctx context.Context) (Payload, bool) {
	payload, ok := ctx.Value(PayloadCtxKey{}).(Payload)
	return payload, ok
}

// GetUserIDFromContext returns subject/user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	payload, ok := GetPayloadFromContext(ctx)
	if !ok {
		return "", false
	}
	return payload.UserID, true
}

// GetUsernameFromContext returns username from context
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	payload, ok := GetPayloadFromContext(ctx)
	if !ok {
		return "", false
	}
	return payload.Username, true
}

// SetScopeToContext attaches Scope to context with trace propagation
func SetScopeToContext(ctx context.Context, scope Scope) context.Context {
	// Ensure trace context is preserved
	tracer := tracing.NewTraceContext()
	if traceID := tracer.GetTraceID(ctx); traceID != "" {
		ctx = tracer.WithTraceID(ctx, traceID)
	}
	return context.WithValue(ctx, ScopeCtxKey{}, scope)
}

// GetScopeFromContext returns Scope from context
func GetScopeFromContext(ctx context.Context) (Scope, bool) {
	scope, ok := ctx.Value(ScopeCtxKey{}).(Scope)
	return scope, ok
}
