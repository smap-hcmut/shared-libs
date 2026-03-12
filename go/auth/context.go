package auth

import (
	"context"

	"github.com/smap/shared-libs/go/tracing"
)

// SetPayloadToContext attaches Payload to context with trace integration.
// This function is compatible with existing service implementations.
func SetPayloadToContext(ctx context.Context, payload Payload) context.Context {
	return context.WithValue(ctx, PayloadCtxKey{}, payload)
}

// GetPayloadFromContext returns Payload from context.
// This function is compatible with existing service implementations.
func GetPayloadFromContext(ctx context.Context) (Payload, bool) {
	payload, ok := ctx.Value(PayloadCtxKey{}).(Payload)
	return payload, ok
}

// GetUserIDFromContext returns subject/user ID from context.
// This function is compatible with existing service implementations.
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	payload, ok := GetPayloadFromContext(ctx)
	if !ok {
		return "", false
	}
	return payload.UserID, true
}

// GetUsernameFromContext returns username from context.
// This function is compatible with existing service implementations.
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	payload, ok := GetPayloadFromContext(ctx)
	if !ok {
		return "", false
	}
	return payload.Username, true
}

// SetScopeToContext attaches Scope to context with trace integration.
// This function is compatible with existing service implementations.
func SetScopeToContext(ctx context.Context, scope Scope) context.Context {
	return context.WithValue(ctx, ScopeCtxKey{}, scope)
}

// GetScopeFromContext returns Scope from context.
// This function is compatible with existing service implementations.
func GetScopeFromContext(ctx context.Context) Scope {
	scope, ok := ctx.Value(ScopeCtxKey{}).(Scope)
	if !ok {
		return Scope{}
	}
	return scope
}

// Enhanced context functions with trace integration

// SetPayloadToContextWithTrace attaches Payload to context and ensures trace_id is available.
// This is an enhanced version that integrates with the tracing system.
func SetPayloadToContextWithTrace(ctx context.Context, payload Payload, tracer tracing.TraceContext) context.Context {
	// Set payload in context
	ctx = SetPayloadToContext(ctx, payload)

	// Ensure trace_id is available - use JWT ID if no trace_id exists
	if tracer.GetTraceID(ctx) == "" && payload.Id != "" {
		// Use JWT ID as trace_id if no trace context exists
		ctx = tracer.WithTraceID(ctx, payload.Id)
	}

	return ctx
}

// SetScopeToContextWithTrace attaches Scope to context and ensures trace_id is available.
// This is an enhanced version that integrates with the tracing system.
func SetScopeToContextWithTrace(ctx context.Context, scope Scope, tracer tracing.TraceContext) context.Context {
	// Set scope in context
	ctx = SetScopeToContext(ctx, scope)

	// Ensure trace_id is available - generate if missing
	if tracer.GetTraceID(ctx) == "" {
		traceID := tracer.GenerateTraceID()
		ctx = tracer.WithTraceID(ctx, traceID)
	}

	return ctx
}

// GetContextInfo returns both auth and trace information from context.
// This is a convenience function for debugging and logging.
func GetContextInfo(ctx context.Context, tracer tracing.TraceContext) ContextInfo {
	payload, hasPayload := GetPayloadFromContext(ctx)
	scope := GetScopeFromContext(ctx)
	traceID := tracer.GetTraceID(ctx)

	return ContextInfo{
		TraceID:    traceID,
		HasPayload: hasPayload,
		Payload:    payload,
		Scope:      scope,
	}
}

// ContextInfo holds combined auth and trace information
type ContextInfo struct {
	TraceID    string  `json:"trace_id"`
	HasPayload bool    `json:"has_payload"`
	Payload    Payload `json:"payload,omitempty"`
	Scope      Scope   `json:"scope,omitempty"`
}
