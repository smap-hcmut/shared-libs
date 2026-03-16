package auth

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// Manager defines the interface for JWT operations with trace integration.
type Manager interface {
	Verify(token string) (Payload, error)
	VerifyWithTrace(ctx context.Context, token string) (Payload, context.Context, error)
	CreateToken(payload Payload) (string, error)
	CreateTokenWithTrace(ctx context.Context, payload Payload) (string, context.Context, error)
	VerifyScope(scopeHeader string) (Scope, error)
}

// JWTManager implements Manager interface for HMAC-based JWT tokens
type JWTManager struct {
	secretKey string
	tracer    tracing.TraceContext
}

// JWKSManager implements Manager interface for JWKS-based JWT tokens
type JWKSManager struct {
	verifier *Verifier
	tracer   tracing.TraceContext
}

// NewManager creates a new HMAC-based JWT manager with trace integration.
func NewManager(secretKey string) Manager {
	return &JWTManager{
		secretKey: secretKey,
		tracer:    tracing.NewTraceContext(),
	}
}

// NewManagerWithTracer creates a new HMAC-based JWT manager with custom tracer.
func NewManagerWithTracer(secretKey string, tracer tracing.TraceContext) Manager {
	return &JWTManager{
		secretKey: secretKey,
		tracer:    tracer,
	}
}

// NewJWKSManager creates a new JWKS-based JWT manager with trace integration.
func NewJWKSManager(verifier *Verifier) Manager {
	return &JWKSManager{
		verifier: verifier,
		tracer:   tracing.NewTraceContext(),
	}
}

// NewJWKSManagerWithTracer creates a new JWKS-based JWT manager with custom tracer.
func NewJWKSManagerWithTracer(verifier *Verifier, tracer tracing.TraceContext) Manager {
	return &JWKSManager{
		verifier: verifier,
		tracer:   tracer,
	}
}

// Verify verifies the JWT token and returns the payload if valid.
// This function is compatible with existing service implementations.
func (m *JWTManager) Verify(token string) (Payload, error) {
	if token == "" {
		return Payload{}, fmt.Errorf("%w: token is empty", ErrInvalidToken)
	}

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrInvalidToken, t.Header["alg"])
		}
		return []byte(m.secretKey), nil
	}

	jwtToken, err := jwt.Parse(token, keyFunc)
	if err != nil {
		return Payload{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !jwtToken.Valid {
		return Payload{}, fmt.Errorf("%w: token is not valid", ErrInvalidToken)
	}

	mapClaims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return Payload{}, fmt.Errorf("%w: failed to parse claims", ErrInvalidToken)
	}

	return payloadFromMapClaims(mapClaims), nil
}

// VerifyWithTrace verifies JWT token and integrates with trace context.
// Returns payload, enhanced context with trace_id, and error.
func (m *JWTManager) VerifyWithTrace(ctx context.Context, token string) (Payload, context.Context, error) {
	payload, err := m.Verify(token)
	if err != nil {
		return payload, ctx, err
	}

	// Enhance context with trace_id from JWT ID if no trace exists
	if m.tracer.GetTraceID(ctx) == "" && payload.Id != "" {
		ctx = m.tracer.WithTraceID(ctx, payload.Id)
	}

	// Set payload in context
	ctx = SetPayloadToContext(ctx, payload)
	ctx = SetScopeToContext(ctx, NewScope(payload))

	return payload, ctx, nil
}

// CreateToken creates a new JWT token with the given payload.
// This function is compatible with existing service implementations.
func (m *JWTManager) CreateToken(payload Payload) (string, error) {
	now := time.Now()
	payload.StandardClaims = jwt.StandardClaims{
		ExpiresAt: now.Add(TokenExpirationDuration).Unix(),
		Id:        fmt.Sprintf("%d", now.UnixNano()),
		NotBefore: now.Unix(),
		IssuedAt:  now.Unix(),
		Subject:   payload.UserID, // set "sub" = UserID so Verify reads it back correctly
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return token.SignedString([]byte(m.secretKey))
}

// CreateTokenWithTrace creates JWT token and integrates with trace context.
// Uses trace_id as JWT ID if available, returns enhanced context.
func (m *JWTManager) CreateTokenWithTrace(ctx context.Context, payload Payload) (string, context.Context, error) {
	now := time.Now()

	// Use trace_id as JWT ID if available
	jwtID := fmt.Sprintf("%d", now.UnixNano())
	if traceID := m.tracer.GetTraceID(ctx); traceID != "" {
		jwtID = traceID
	}

	payload.StandardClaims = jwt.StandardClaims{
		ExpiresAt: now.Add(TokenExpirationDuration).Unix(),
		Id:        jwtID,
		NotBefore: now.Unix(),
		IssuedAt:  now.Unix(),
		Subject:   payload.UserID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", ctx, err
	}

	// Ensure trace_id is in context
	if m.tracer.GetTraceID(ctx) == "" {
		ctx = m.tracer.WithTraceID(ctx, jwtID)
	}

	// Set payload in context
	ctx = SetPayloadToContext(ctx, payload)
	ctx = SetScopeToContext(ctx, NewScope(payload))

	return tokenString, ctx, nil
}

// VerifyScope parses the scope header and returns the scope.
// This function is compatible with existing service implementations.
func (m *JWTManager) VerifyScope(scopeHeader string) (Scope, error) {
	return ParseScopeHeader(scopeHeader)
}

// JWKS Manager methods

// Verify verifies the JWT token using JWKS and returns the payload if valid.
func (m *JWKSManager) Verify(token string) (Payload, error) {
	return m.verifier.VerifyToken(token)
}

// VerifyWithTrace verifies JWT token using JWKS and integrates with trace context.
func (m *JWKSManager) VerifyWithTrace(ctx context.Context, token string) (Payload, context.Context, error) {
	payload, err := m.Verify(token)
	if err != nil {
		return payload, ctx, err
	}

	// Enhance context with trace_id from JWT ID if no trace exists
	if m.tracer.GetTraceID(ctx) == "" && payload.Id != "" {
		ctx = m.tracer.WithTraceID(ctx, payload.Id)
	}

	// Set payload in context
	ctx = SetPayloadToContext(ctx, payload)
	ctx = SetScopeToContext(ctx, NewScope(payload))

	return payload, ctx, nil
}

// CreateToken creates a new JWT token (not supported for JWKS manager).
func (m *JWKSManager) CreateToken(payload Payload) (string, error) {
	return "", fmt.Errorf("token creation not supported for JWKS manager")
}

// CreateTokenWithTrace creates JWT token (not supported for JWKS manager).
func (m *JWKSManager) CreateTokenWithTrace(ctx context.Context, payload Payload) (string, context.Context, error) {
	return "", ctx, fmt.Errorf("token creation not supported for JWKS manager")
}

// VerifyScope parses the scope header and returns the scope.
func (m *JWKSManager) VerifyScope(scopeHeader string) (Scope, error) {
	return ParseScopeHeader(scopeHeader)
}

// Helper functions compatible with existing service implementations

func payloadFromMapClaims(claims jwt.MapClaims) Payload {
	payload := Payload{
		UserID:   getStringClaim(claims, "sub"),
		Username: firstNonEmptyClaim(claims, "username", "email"),
		Role:     getStringClaim(claims, "role"),
		Type:     getStringClaim(claims, "type"),
		Refresh:  getBoolClaim(claims, "refresh"),
	}
	payload.StandardClaims = jwt.StandardClaims{
		Audience:  getAudienceClaim(claims),
		ExpiresAt: getInt64Claim(claims, "exp"),
		Id:        firstNonEmptyClaim(claims, "jti", "id"),
		IssuedAt:  getInt64Claim(claims, "iat"),
		Issuer:    getStringClaim(claims, "iss"),
		NotBefore: getInt64Claim(claims, "nbf"),
		Subject:   getStringClaim(claims, "sub"),
	}
	return payload
}

func getStringClaim(claims jwt.MapClaims, key string) string {
	value, ok := claims[key]
	if !ok || value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func firstNonEmptyClaim(claims jwt.MapClaims, keys ...string) string {
	for _, key := range keys {
		value := getStringClaim(claims, key)
		if value != "" {
			return value
		}
	}
	return ""
}

func getInt64Claim(claims jwt.MapClaims, key string) int64 {
	value, ok := claims[key]
	if !ok || value == nil {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	case string:
		n, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return n
		}
	}
	return 0
}

func getBoolClaim(claims jwt.MapClaims, key string) bool {
	value, ok := claims[key]
	if !ok || value == nil {
		return false
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		parsed, err := strconv.ParseBool(v)
		return err == nil && parsed
	default:
		return false
	}
}

func getAudienceClaim(claims jwt.MapClaims) string {
	value, ok := claims["aud"]
	if !ok || value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				return s
			}
		}
	case []string:
		for _, item := range v {
			if item != "" {
				return item
			}
		}
	}
	return ""
}
